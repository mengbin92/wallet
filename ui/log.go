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
	// 启动 goroutine 持续更新 UI
	go func() {
		for msg := range lw.ch {
			entry.SetText(entry.Text + msg)
			entry.CursorRow = len(entry.Text)
			entry.Refresh()
		}
	}()
	return lw
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	msg := string(p)
	fyne.DoAndWait(func() {
		lw.entry.SetText(lw.entry.Text + msg)
		lw.entry.CursorRow = len(lw.entry.Text)
		lw.entry.Refresh()
	})
	return len(p), nil
}

