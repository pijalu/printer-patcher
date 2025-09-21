package tools

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type WidgetWriter struct {
	w *widget.TextGrid
}

func (ww *WidgetWriter) Write(p []byte) (n int, err error) {
	out := string(p[:])
	fyne.DoAndWait(func() {
		ww.w.Append(out)
	})
	return len(p), nil
}

func NewWidgetWriter(w *widget.TextGrid) *WidgetWriter {
	return &WidgetWriter{
		w: w,
	}
}
