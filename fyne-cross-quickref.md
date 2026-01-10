# fyne-cross 快速参考

## 常用命令

### 发布版本编译（隐藏 test/dev 配置）

```bash
# Windows
fyne-cross windows -arch=amd64 -tags release -app-id com.example.cryptotool

# macOS
fyne-cross darwin -arch=amd64 -tags release -app-id com.example.cryptotool

# Linux
fyne-cross linux -arch=amd64 -tags release -app-id com.example.cryptotool
```

### 调试版本编译（显示 test/dev 配置）

```bash
# Windows（不传递 tags）
fyne-cross windows -arch=amd64 -app-id com.example.cryptotool

# macOS
fyne-cross darwin -arch=amd64 -app-id com.example.cryptotool

# Linux
fyne-cross linux -arch=amd64 -app-id com.example.cryptotool
```

## 重要参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-tags release` | 发布模式（隐藏 test/dev） | 必须传递此参数 |
| `-app-id` | 应用程序包名 | com.example.cryptotool |
| `-arch` | 目标架构 | amd64, arm64 |
| `-name` | 应用程序名称 | 默认使用目录名 |

## 使用 Make 简化命令

```bash
# 编译所有平台（推荐）
make fyne-release

# 单个平台
make fyne-release-windows
make fyne-release-macos
make fyne-release-linux
```

## 输出目录

编译完成后的文件位于：
```
fyne-cross/dist/
├── windows-amd64/    # Windows .exe
├── darwin-amd64/     # macOS .app
└── linux-amd64/      # Linux 可执行文件
```

## 常见问题

### Q: 如何修改应用程序图标？
A: 创建 `Icon.png` 文件在项目根目录，fyne-cross 会自动使用。

### Q: 如何验证是否为发布版本？
A:
```bash
strings fyne-cross/dist/windows-amd64/wallet_ui.exe | grep '"dev"'
# 无输出 = 发布版本 ✅
# 有输出 = 调试版本 ❌
```

### Q: 如何指定应用程序名称？
A: 使用 `-name` 参数：
```bash
fyne-cross windows -arch=amd64 -name "MyWallet" -tags release
```

### Q: 如何编译 ARM64 版本？
A:
```bash
fyne-cross windows -arch=arm64 -tags release
fyne-cross darwin -arch=arm64 -tags release
```

## 发布检查清单

在发布前，请确认：

- ✅ 传递了 `-tags release`
- ✅ 设置了正确的 `-app-id`
- ✅ 在目标平台上测试了生成的安装包
- ✅ 验证不包含 test/dev 配置
- ✅ 检查应用程序名称和图标正确

## 技术说明

本方案使用 **Go build tags** 来控制编译模式，避免 `-ldflags` 参数解析问题：

- **调试模式**: 默认，不传递 `-tags`
- **发布模式**: 传递 `-tags release`
