package main

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type LogWriter struct {
	rt *widget.RichText
	mu sync.Mutex
	ch chan string
}

func NewLogWriter(rt *widget.RichText) *LogWriter {
	lw := &LogWriter{
		rt: rt,
		ch: make(chan string, 100),
	}

	go func() {
		for msg := range lw.ch {
			rt.Segments = append(rt.Segments, &widget.TextSegment{
				Text: msg,
			})
			rt.Refresh()
		}
	}()

	return lw
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
