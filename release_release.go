//go:build release

// +build release

package main

func init() {
	// 发布模式：隐藏 test/dev 配置
	debugMode = "0"
}
