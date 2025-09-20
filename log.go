package main

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type LogWriter struct {
	rt *widget.RichText
	mu sync.Mutex
}

func NewLogWriter(rt *widget.RichText) *LogWriter {
	return &LogWriter{rt: rt}
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	msg := string(p)
	fyne.DoAndWait(func() {
		lw.rt.Segments = append(lw.rt.Segments, &widget.TextSegment{
			Text: msg,
		})
		lw.rt.Refresh()
	})
	return len(p), nil
}
