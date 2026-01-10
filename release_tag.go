//go:build !release

// +build !release

package main

func init() {
	// 调试模式：显示 test/dev 配置
	debugMode = "1"
}
