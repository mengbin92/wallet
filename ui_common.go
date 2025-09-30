package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

var DefaultWindowSize = fyne.NewSize(600, 400)

// NewLogBox 创建日志框
// NewLogBox 创建日志框（固定高度，可滚动）
// NewLogBox 创建日志框（可复制内容）
func NewLogBox(height float32) (*container.Scroll, *widget.Entry) {
    entry := widget.NewMultiLineEntry()
    entry.SetPlaceHolder("日志输出...")
    entry.Wrapping = fyne.TextWrapWord
	entry.OnChanged = func(s string) {
	}


    scroll := container.NewScroll(entry)
    scroll.SetMinSize(fyne.NewSize(400, height))

    return scroll, entry
}


// NewFolderButton 浏览文件夹按钮
func NewFolderButton(entry *widget.Entry, w fyne.Window) *widget.Button {
	return widget.NewButton("浏览", func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("获取Home目录失败:", err)
			return
		}

		uri := storage.NewFileURI(filepath.Clean(homeDir))
		listable, err := storage.ListerForURI(uri)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		fd := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if lu == nil {
				return
			}
			entry.SetText(lu.Path())
		}, w)

		fd.SetLocation(listable)
		fd.Show()
	})
}

// AppDir 获取可执行文件所在目录
func AppDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}
