package main

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type LogWriter struct {
	entry *widget.Entry
	mu    sync.Mutex
	ch    chan string
}

func NewLogWriter(entry *widget.Entry) *LogWriter {
	lw := &LogWriter{
		entry: entry,
		ch:    make(chan string, 100),
	}

	// 异步追加日志
	go func() {
		for msg := range lw.ch {
			fyne.DoAndWait(func() {
				lw.entry.SetText(lw.entry.Text + msg) // 直接拼接
				lw.entry.CursorRow = len(lw.entry.Text) // 保持光标在最后（方便滚动）
				lw.entry.Refresh()
			})
		}
	}()

	return lw
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	msg := string(p)
	lw.ch <- msg // 通过 channel 异步写入
	return len(p), nil
}
