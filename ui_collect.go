package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/mengbin92/wallet/internal/collector"
	"github.com/mengbin92/wallet/internal/config"
)

// 开发模式：设置为 "1" 时显示 test 和 dev 配置，"0" 时隐藏
// 通过 Go build tags 控制：
// - 默认（调试模式）：显示 test/dev
// - -tags release：隐藏 test/dev
var debugMode string // 由 release_tag.go 或 release_release.go 初始化

// CollectionStats 用于跟踪归集统计信息
type CollectionStats struct {
	mu                sync.Mutex
	TotalAccounts     int
	ProcessedAccounts int
	SuccessCount      int
	FailureCount      int
	TotalAmount       string
	StartTime         time.Time
}

// 全局统计变量（用于在闭包中共享）
var globalStats = &CollectionStats{}

func NewCollectTab(w fyne.Window) *container.TabItem {
	rpcEntry := widget.NewEntry()
	rpcEntry.SetText("https://public-bsc-mainnet.fastnode.io")

	mainKeyEntry := widget.NewEntry()
	mainKeyEntry.SetPlaceHolder("请输入主账户的目录...")
	mainKeyBtn := NewFolderButton(mainKeyEntry, w)
	mainKeyRow := container.NewBorder(nil, nil, nil, mainKeyBtn, mainKeyEntry)

	ksEntry := widget.NewEntry()
	ksEntry.SetPlaceHolder("请选择存放keystore的目录...")
	browseBtn := NewFolderButton(ksEntry, w)
	row := container.NewBorder(nil, nil, nil, browseBtn, ksEntry)

	// Keystore密码输入
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("请输入keystore密码...")
	passwordEntry.SetText("") // 默认为空，用户可输入

	tokenEntry := widget.NewEntry()
	toEntry := widget.NewEntry()

	// ========== 新增：配置选项 ==========

	// 并发模式选择
	enableConcurrent := widget.NewCheck("启用并发模式 (5-20倍性能提升)", func(checked bool) {})
	enableConcurrent.Checked = true

	// Worker数量选择
	workerSlider := widget.NewSlider(1, 20)
	workerSlider.Value = 10
	workerSlider.Step = 1
	workerLabel := widget.NewLabel("Worker数量: 10")
	workerSlider.OnChanged = func(value float64) {
		workerLabel.SetText("Worker数量: " + strconv.Itoa(int(value)))
	}

	// 配置预设选择 - 根据调试模式决定显示哪些选项
	var configOptions []string
	if debugMode == "1" {
		configOptions = []string{"default", "prod", "test", "dev"}
	} else {
		configOptions = []string{"default", "prod"}
	}

	configPreset := widget.NewSelect(configOptions, nil)
	configPreset.SetSelected("default")
	configPreset.OnChanged = func(s string) {
		// 根据预设自动调整worker数量
		switch s {
		case "default":
			workerSlider.SetValue(5)
		case "prod":
			workerSlider.SetValue(10)
		case "test":
			workerSlider.SetValue(2)
		case "dev":
			workerSlider.SetValue(1)
		}
	}

	// 配置选项容器
	configBoxItems := []fyne.CanvasObject{
		widget.NewSeparator(),
		widget.NewLabel("⚙️ 并发配置 (可选)"),
		container.NewGridWithColumns(2,
			widget.NewLabel("配置预设:"),
			configPreset,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("Worker数量:"),
			container.NewBorder(nil, nil, nil, workerLabel, workerSlider),
		),
		widget.NewLabel("💡 提示: 10账户以下用5 workers, 10-50账户用10 workers, 50+账户用15-20 workers"),
	}

	// 只在调试模式下显示 dev 模式提示
	if debugMode == "1" {
		configBoxItems = append(configBoxItems, widget.NewLabel("🔧 dev模式适用于 geth -dev 本地测试网 (0确认, 10秒超时)"))
	}

	configBox := container.NewVBox(configBoxItems...)

	// ========== 进度显示 ==========

	// 进度条
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// 统计信息标签
	statsLabel := widget.NewLabel("等待开始...")
	statsLabel.Hide()

	// 当前处理标签
	currentLabel := widget.NewLabel("")
	currentLabel.Hide()

	scroll, rt := NewLogBox(200)
	lw := NewLogWriter(rt)
	logger := log.New(lw, "[Collect] ", log.LstdFlags)

	var collectBtn *widget.Button
	collectBtn = widget.NewButton("开始归集", func() {
		go func() {
			// 禁用按钮，防止重复点击
			collectBtn.Disable()
			defer collectBtn.Enable()

			logger.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			logger.Println("🚀 开始执行归集...")

			// 验证必填参数
			if rpcEntry.Text == "" || mainKeyEntry.Text == "" || ksEntry.Text == "" || tokenEntry.Text == "" || toEntry.Text == "" {
				logger.Println("❌ 请填写完整参数")
				return
			}

			// 显示进度UI
			progressBar.Show()
			statsLabel.Show()
			currentLabel.Show()

			// 重置进度和全局统计
			progressBar.SetValue(0)
			globalStats = &CollectionStats{StartTime: time.Now()}

			var err error

			// 判断是否使用并发模式
			if enableConcurrent.Checked {
				// 使用并发归集
				logger.Println("⚡ 使用并发归集模式")

				// 加载配置
				var cfg *config.Config
				switch configPreset.Selected {
				case "prod", "production":
					cfg = config.ProdConfig()
					logger.Println("📋 使用生产配置 (10 workers, 5次重试)")
				case "test":
					cfg = config.TestConfig()
					logger.Println("📋 使用测试配置 (2 workers, 快速确认)")
				case "dev":
					cfg = config.DevConfig()
					logger.Println("📋 使用开发配置 (geth -dev模式, 0确认, 10秒超时)")
				default:
					cfg = config.DefaultConfig()
					logger.Println("📋 使用默认配置 (5 workers, 3次重试)")
				}

				// 覆盖worker数量
				workers := int(workerSlider.Value)
				cfg.Concurrency.MaxWorkers = workers
				logger.Printf("⚙️  Worker数量: %d", workers)
				logger.Printf("⚡ 预期性能提升: %dx-%dx\n", workers/2, workers)

				// 创建一个包装的logger来捕获进度
				progressLogger := &progressLogger{
					Logger:     logger,
					stats:      globalStats,
					progress:  progressBar,
					statsLabel: statsLabel,
					current:    currentLabel,
				}

				// 在后台启动日志监听
				logChan := make(chan string, 100)
				done := make(chan bool)

				go func() {
					for msg := range logChan {
						progressLogger.updateProgress(msg)
					}
					done <- true
				}()

				// 创建自定义writer来捕获日志
				logWriter := &logWriter{
					ch: logChan,
				}

				// 包装logger以捕获输出
				wrappedLogger := log.New(io.MultiWriter(lw, logWriter), "[Collect] ", log.LstdFlags)

				err = collector.CollectTokensConcurrent(
					rpcEntry.Text,
					mainKeyEntry.Text,
					ksEntry.Text,
					passwordEntry.Text,
					tokenEntry.Text,
					toEntry.Text,
					cfg,
					wrappedLogger,
				)

				// 等待所有日志处理完成后再关闭channel
				close(logChan)
				<-done // 等待goroutine处理完所有消息
			} else {
				// 使用串行归集（兼容模式）
				logger.Println("🔄 使用串行归集模式 (兼容模式)")
				logger.Println("⚠️  提示: 启用并发模式可获得5-20倍性能提升\n")

				err = collector.CollectTokens(
					rpcEntry.Text,
					mainKeyEntry.Text,
					ksEntry.Text,
					passwordEntry.Text,
					tokenEntry.Text,
					toEntry.Text,
					logger,
				)
			}

			if err != nil {
				logger.Println("❌ 代币归集出错:", err.Error())
				currentLabel.SetText("错误: " + err.Error())
				return
			}

			elapsed := time.Since(globalStats.StartTime)
			logger.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			logger.Printf("✅ 代币归集完成! 总耗时: %v\n", elapsed)
		}()
	})

	content := container.NewVBox(
		widget.NewLabel("代币归集"),
		widget.NewForm(
			widget.NewFormItem("RPC", rpcEntry),
			widget.NewFormItem("主账户目录", mainKeyRow),
			widget.NewFormItem("Keystore", row),
			// widget.NewFormItem("Keystore密码", passwordEntry),
			widget.NewFormItem("Token合约地址", tokenEntry),
			widget.NewFormItem("接收地址", toEntry),
		),
		enableConcurrent,
		configBox,
		collectBtn,
		progressBar,
		currentLabel,
		statsLabel,
		scroll,
	)

	return container.NewTabItem("代币归集", content)
}

// logWriter 自定义writer用于捕获日志
type logWriter struct {
	ch chan string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.ch <- string(p)
	return len(p), nil
}

// progressLogger 包装标准logger以捕获进度信息
type progressLogger struct {
	*log.Logger // 嵌入标准logger
	stats      *CollectionStats
	progress   *widget.ProgressBar
	statsLabel *widget.Label
	current    *widget.Label
}

func (p *progressLogger) Printf(format string, v ...interface{}) {
	// 调用嵌入的logger
	p.Logger.Printf(format, v...)

	// 解析日志以提取进度信息
	msg := fmt.Sprintf(format, v...)
	p.updateProgress(msg)
}

func (p *progressLogger) Println(v ...interface{}) {
	p.Logger.Println(v...)

	// 转换为字符串并更新进度
	msg := fmt.Sprint(v...)
	p.updateProgress(msg)
}

func (p *progressLogger) updateProgress(msg string) {
	// 清理消息：移除时间戳前缀 [Collect] 2025/01/10 10:30:45
	cleanMsg := msg
	if idx := strings.Index(msg, "] "); idx >= 0 && idx < 20 {
		cleanMsg = msg[idx+2:]
	}

	// 检测账户总数 - 解析 "找到 X 个有余额的账户"
	if strings.Contains(cleanMsg, "找到") && strings.Contains(cleanMsg, "个有余额的账户") {
		p.stats.mu.Lock()
		defer p.stats.mu.Unlock()
		// 从日志中提取账户数量
		parts := strings.Fields(cleanMsg)
		for _, part := range parts {
			if num, err := strconv.Atoi(part); err == nil {
				p.stats.TotalAccounts = num
				p.current.SetText(cleanMsg)
				p.progress.SetValue(0) // 重置进度条
				fmt.Printf("[DEBUG] TotalAccounts set to: %d\n", num) // 调试输出
				return
			}
		}
		p.current.SetText(cleanMsg)
	}

	// 检测处理进度 - "开始归集" 表示真正开始处理一个账户
	if strings.Contains(cleanMsg, "开始归集") {
		p.stats.mu.Lock()
		p.stats.ProcessedAccounts++
		current := float64(p.stats.ProcessedAccounts)

		fmt.Printf("[DEBUG] ProcessedAccounts: %d/%d\n", p.stats.ProcessedAccounts, p.stats.TotalAccounts) // 调试输出

		if p.stats.TotalAccounts > 0 {
			progress := current / float64(p.stats.TotalAccounts)
			fmt.Printf("[DEBUG] Progress set to: %.2f%%\n", progress*100) // 调试输出
			// 使用 goroutine 来更新 UI，避免阻塞日志处理
			go p.progress.SetValue(progress)
		}

		p.current.SetText(cleanMsg)
		p.stats.mu.Unlock()
	}

	// "发现余额" 只用于显示，不更新进度
	if strings.Contains(cleanMsg, "发现余额") {
		p.current.SetText(cleanMsg)
	}

	// 检测成功 - "归集 ... 成功" 而不是所有包含"成功"的消息
	if strings.Contains(cleanMsg, "归集") && strings.Contains(cleanMsg, "成功") && !strings.Contains(cleanMsg, "失败") {
		p.stats.mu.Lock()
		p.stats.SuccessCount++
		p.updateStatsLabel()
		p.stats.mu.Unlock()
	}

	// 检测失败 - "归集 ... 失败" 而不是所有包含"失败"的消息
	if (strings.Contains(cleanMsg, "归集") || strings.Contains(cleanMsg, "获取")) && strings.Contains(cleanMsg, "失败") {
		p.stats.mu.Lock()
		p.stats.FailureCount++
		p.updateStatsLabel()
		p.stats.mu.Unlock()
	}

	// 检测统计信息
	if strings.Contains(cleanMsg, "归集统计") || strings.Contains(cleanMsg, "成功率为") {
		p.statsLabel.SetText(cleanMsg)
	}
}

func (p *progressLogger) updateStatsLabel() {
	p.statsLabel.SetText(fmt.Sprintf(
		"已处理: %d | 成功: %d | 失败: %d",
		p.stats.ProcessedAccounts,
		p.stats.SuccessCount,
		p.stats.FailureCount,
	))
}
