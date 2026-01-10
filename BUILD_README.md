# 编译说明

## 调试模式 vs 发布模式

本项目支持两种编译模式：

### 调试模式（默认）
- **显示所有配置选项**：default, prod, test, dev
- **显示 Dev 模式提示**
- **用途**：开发和测试时使用

### 发布模式
- **仅显示生产配置**：default, prod
- **隐藏 test 和 dev 配置**
- **隐藏 Dev 模式提示**
- **用途**：发布给生产用户使用

## 编译方式

### 方式1：使用 fyne-cross（推荐）

`fyne-cross` 是 Fyne 官方提供的跨平台编译工具，可以生成独立的安装包。

#### 安装 fyne-cross
```bash
go install github.com/fyne-io/fyne-cross/cmd/fyne-cross@latest
```

#### 使用 Make 编译（推荐）
```bash
# 查看所有可用命令
make help

# 编译所有平台的发布版本
make fyne-release

# 编译单个平台
make fyne-release-windows  # Windows
make fyne-release-macos    # macOS
make fyne-release-linux    # Linux
```

#### 直接使用 fyne-cross 命令
```bash
# Windows 发布版本
fyne-cross windows -arch=amd64 -tags release -app-id com.example.cryptotool

# macOS 发布版本
fyne-cross darwin -arch=amd64 -tags release -app-id com.example.cryptotool

# Linux 发布版本
fyne-cross linux -arch=amd64 -tags release -app-id com.example.cryptotool
```

**重要参数说明：**
- `-tags release` → 设置为发布模式（隐藏 test/dev 配置）
- `-app-id com.example.cryptotool` → 应用程序 ID（可根据需要修改）

#### 编译输出
编译完成后，独立安装包位于 `fyne-cross/dist/` 目录：
- Windows: `fyne-cross/dist/windows-amd64/`
- macOS: `fyne-cross/dist/darwin-amd64/`
- Linux: `fyne-cross/dist/linux-amd64/`

### 方式2：本地编译（用于开发测试）

#### 调试模式
```bash
go build -o wallet_ui .
# 或
make
```
配置选项：`default`, `prod`, `test`, `dev`

#### 发布模式
```bash
go build -tags release -o wallet_ui .
# 或
make release
```
配置选项：`default`, `prod`

### 方式3：传统 Go 跨平台编译（不推荐）

用于快速测试，但生成的不是独立安装包：

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -tags release -o wallet_ui.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -tags release -o wallet_ui-macos .

# Linux
GOOS=linux GOARCH=amd64 go build -tags release -o wallet_ui-linux .
```

## 常用命令

```bash
# 本地开发编译（调试模式）
make

# 本地发布编译
make release

# 使用 fyne-cross 编译所有平台（推荐）
make fyne-release

# 清理编译文件
make clean

# 运行测试
make test
```

## 配置说明

### default（默认配置）
- **Workers**: 5
- **确认数**: 6
- **重试次数**: 3
- **适用场景**: 10账户以下的小规模归集

### prod（生产配置）
- **Workers**: 10
- **确认数**: 6
- **重试次数**: 5
- **适用场景**: 生产环境，10-50账户

### test（测试配置）- 仅调试模式
- **Workers**: 2
- **确认数**: 6
- **超时时间**: 30秒
- **适用场景**: 测试环境

### dev（开发配置）- 仅调试模式
- **Workers**: 1
- **确认数**: 0
- **超时时间**: 10秒
- **适用场景**: geth -dev 本地开发网络
- **特点**: 无需等待出块，交易进入内存池即可

## 验证编译模式

编译后可以通过以下方式验证：

```bash
# 检查是否包含 dev 配置（调试模式会显示，发布模式不会）
strings wallet_ui | grep '"dev"'
```

- **有输出** → 调试版本（包含 test/dev）
- **无输出** → 发布版本（仅 default/prod）

## 发布检查清单

发布前请确认：

- [ ] 使用 `make fyne-release` 或 `fyne-cross` 编译
- [ ] 传递了 `-tags release` 参数
- [ ] 设置了正确的 `-app-id`
- [ ] 验证编译后的二进制文件不包含 test/dev 配置
- [ ] 测试安装包在目标平台上能正常运行

## 技术实现说明

本方案使用 **Go build tags** 来控制编译模式：

- **默认（调试模式）**: 编译 `release_tag.go`，设置 `debugMode = "1"`
- **`-tags release`**: 编译 `release_release.go`，设置 `debugMode = "0"`

### 相关文件

- `ui_collect.go` - 主文件，声明 `debugMode` 变量
- `release_tag.go` - 调试模式配置（默认）
- `release_release.go` - 发布模式配置（需 `-tags release`）

