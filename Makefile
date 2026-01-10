.PHONY: all build release clean test help fyne-release fyne-release-windows fyne-release-macos fyne-release-linux

# 默认目标：调试模式（显示 test 和 dev 配置）
all: build

# 调试版本：显示所有配置选项（包括 test 和 dev）
build:
	@echo "编译本地调试版本（包含 test 和 dev 配置）..."
	@go build -o wallet_ui .
	@echo "✅ 调试版本编译完成: ./wallet_ui"

# 发布版本：隐藏 test 和 dev 配置
release:
	@echo "编译本地发布版本（仅包含 default 和 prod 配置）..."
	@go build -tags release -o wallet_ui .
	@echo "✅ 发布版本编译完成: ./wallet_ui"

# ========== 使用 fyne-cross 编译独立应用 ==========

# 使用 fyne-cross 编译所有平台（发布模式）
fyne-release: fyne-release-windows fyne-release-macos fyne-release-linux
	@echo "✅ 所有平台发布版本编译完成！"

# Windows 发布版本（使用 fyne-cross）
fyne-release-windows:
	@echo "编译 Windows 发布版本（使用 fyne-cross）..."
	@fyne-cross windows -arch=amd64 -tags release -app-id com.example.cryptotool
	@echo "✅ Windows 发布版本编译完成"

# macOS 发布版本（使用 fyne-cross）
fyne-release-macos:
	@echo "编译 macOS 发布版本（使用 fyne-cross）..."
	@fyne-cross darwin -arch=amd64 -tags release -app-id com.example.cryptotool
	@echo "✅ macOS 发布版本编译完成"

# Linux 发布版本（使用 fyne-cross）
fyne-release-linux:
	@echo "编译 Linux 发布版本（使用 fyne-cross）..."
	@fyne-cross linux -arch=amd64 -tags release -app-id com.example.cryptotool
	@echo "✅ Linux 发布版本编译完成"

# ========== 传统的 Go 编译（不推荐，用于测试） ==========

# Windows 发布版本（传统方式，不推荐）
release-windows-old:
	@echo "编译 Windows 发布版本（传统方式，不推荐）..."
	@GOOS=windows GOARCH=amd64 go build -tags release -o wallet_ui.exe .
	@echo "✅ Windows 发布版本编译完成: ./wallet_ui.exe"

# macOS 发布版本（传统方式，不推荐）
release-macos-old:
	@echo "编译 macOS 发布版本（传统方式，不推荐）..."
	@go build -tags release -o wallet_ui-macos .
	@echo "✅ macOS 发布版本编译完成: ./wallet_ui-macos"

# Linux 发布版本（传统方式，不推荐）
release-linux-old:
	@echo "编译 Linux 发布版本（传统方式，不推荐）..."
	@GOOS=linux GOARCH=amd64 go build -tags release -o wallet_ui-linux .
	@echo "✅ Linux 发布版本编译完成: ./wallet_ui-linux"

# 清理
clean:
	@echo "清理编译文件..."
	@rm -f wallet_ui wallet_ui.exe wallet_ui-macos wallet_ui-linux
	@rm -rf fyne-cross/dist/*
	@echo "✅ 清理完成"

# 测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 帮助信息
help:
	@echo "可用命令:"
	@echo ""
	@echo "本地编译:"
	@echo "  make              - 编译本地调试版本（包含 test/dev 配置）"
	@echo "  make build        - 同上"
	@echo "  make release      - 编译本地发布版本（仅 default/prod 配置）"
	@echo ""
	@echo "使用 fyne-cross 编译独立应用（推荐）:"
	@echo "  make fyne-release          - 编译所有平台发布版本"
	@echo "  make fyne-release-windows  - 编译 Windows 发布版本"
	@echo "  make fyne-release-macos    - 编译 macOS 发布版本"
	@echo "  make fyne-release-linux    - 编译 Linux 发布版本"
	@echo ""
	@echo "传统 Go 编译（不推荐，仅用于测试）:"
	@echo "  make release-windows-old   - 编译 Windows 发布版本"
	@echo "  make release-macos-old     - 编译 macOS 发布版本"
	@echo "  make release-linux-old     - 编译 Linux 发布版本"
	@echo ""
	@echo "其他:"
	@echo "  make clean         - 清理编译文件"
	@echo "  make test          - 运行测试"
	@echo "  make help          - 显示此帮助信息"
	@echo ""
	@echo "说明:"
	@echo "  - 调试模式：显示 default, prod, test, dev 配置"
	@echo "  - 发布模式：仅显示 default, prod 配置"
	@echo "  - 推荐使用 fyne-cross 编译独立应用程序"
