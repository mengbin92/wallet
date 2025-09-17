package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mengbin92/wallet/internal/collector"
	"github.com/mengbin92/wallet/internal/wallet"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Crypto Tool UI")

	// --- 创建钱包 UI ---
	countEntry := widget.NewEntry()
	countEntry.SetPlaceHolder("数量")
	passEntry := widget.NewPasswordEntry()
	chainEntry := widget.NewEntry()
	chainEntry.SetText("eth")
	outEntry := widget.NewEntry()
	outEntry.SetText("./keystore")
	logBox1 := widget.NewMultiLineEntry()
	logBox1.SetMinRowsVisible(10)
	lw1 := NewLogWriter(logBox1)
	walletLogger := log.New(lw1, "[Wallet] ", log.LstdFlags)

	createBtn := widget.NewButton("生成钱包", func() {
		go func() {
			walletLogger.Println("开始执行钱包生成...")
			// logBox1.SetText("开始生成钱包...\n")
			count, err := strconv.Atoi(countEntry.Text)
			if err != nil {
				walletLogger.Println("数量转换出错:", err.Error())
				// logBox1.SetText(logBox1.Text + "错误: " + err.Error() + "\n")
				return
			}
			if countEntry.Text == "" || chainEntry.Text == "" || outEntry.Text == "" {
				walletLogger.Println("请填写完整参数")
				// logBox1.SetText(logBox1.Text + "错误: 请填写完整参数\n")
				return
			}

			addresses, err := wallet.DeriveBatchEVM(chainEntry.Text, "", countEntry.Text, count)
			if err != nil {
				walletLogger.Println("地址生成出错:", err.Error())
				// logBox1.SetText(logBox1.Text + "地址生成出错: " + err.Error() + "\n")
				return
			}
			walletLogger.Println("地址生成完成")
			// logBox1.SetText(logBox1.Text + "地址生成完成!\n")
			for _, address := range addresses {
				walletLogger.Println("地址: " + address.Address)
			}

			err = wallet.SaveToKeystore(addresses, "", filepath.Join(getAppDir(), outEntry.Text))
			if err != nil {
				walletLogger.Println("keystore存储出错:", err.Error())
				// logBox1.SetText(logBox1.Text + "keystore存储出错: " + err.Error() + "\n")
				return
			}
			walletLogger.Println("keystore存储完成")
			// logBox1.SetText(logBox1.Text + "keystore存储完成!\n")
		}()
	})

	walletTab := container.NewVBox(
		widget.NewLabel("钱包创建"),
		widget.NewForm(
			widget.NewFormItem("数量", countEntry),
			widget.NewFormItem("密码", passEntry),
			widget.NewFormItem("底链", chainEntry),
			widget.NewFormItem("输出目录", outEntry),
		),
		createBtn,
		logBox1,
	)

	// --- 归集 UI ---
	rpcEntry := widget.NewEntry()
	rpcEntry.SetText("http://127.0.0.1:8545")
	ksEntry := widget.NewEntry()
	ksEntry.SetText("./keystore")
	// pass2Entry := widget.NewPasswordEntry()
	tokenEntry := widget.NewEntry()
	toEntry := widget.NewEntry()
	logBox2 := widget.NewMultiLineEntry()
	logBox2.SetMinRowsVisible(10)
	lw2 := NewLogWriter(logBox2)
	collectLogger := log.New(lw2, "[Collect] ", log.LstdFlags)

	collectBtn := widget.NewButton("开始归集", func() {
		go func() {
			collectLogger.Println("开始执行归集...")
			// logBox2.SetText("正在执行归集...\n")
			if rpcEntry.Text == "" || ksEntry.Text == "" || tokenEntry.Text == "" || toEntry.Text == "" {
				collectLogger.Println("请填写完整参数")
				// logBox2.SetText(logBox2.Text + "错误: 请填写完整参数\n")
				return
			}
			err := collector.CollectTokens(rpcEntry.Text, filepath.Join(getAppDir(), ksEntry.Text), "", tokenEntry.Text, toEntry.Text, collectLogger)
			if err != nil {
				collectLogger.Println("代币归集出错:", err.Error())
				// logBox2.SetText(logBox2.Text + "代币归集出错: " + err.Error() + "\n")
				return
			} else {
				collectLogger.Println("代币归集完成")
				// logBox2.SetText(logBox2.Text + "完成!")
			}
		}()
	})

	collectTab := container.NewVBox(
		widget.NewLabel("代币归集"),
		widget.NewForm(
			widget.NewFormItem("RPC", rpcEntry),
			widget.NewFormItem("Keystore", ksEntry),
			// widget.NewFormItem("密码", pass2Entry),
			widget.NewFormItem("Token合约地址", tokenEntry),
			widget.NewFormItem("接收地址", toEntry),
		),
		collectBtn,
		logBox2,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("钱包创建", walletTab),
		container.NewTabItem("代币归集", collectTab),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(500, 400))
	w.ShowAndRun()
}

func getAppDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}
