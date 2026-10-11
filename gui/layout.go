package gui

import (
	"fyne.io/fyne/v2"
)

// adaptiveGridLayout arranges its objects in one column on narrow widths and
// two columns once the available width reaches the threshold. Row heights are
// driven by the tallest object in each row.
type adaptiveGridLayout struct {
	threshold float32
}

func newAdaptiveGridLayout() fyne.Layout {
	return &adaptiveGridLayout{threshold: 560}
}

func (g *adaptiveGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	var w, h float32
	for _, o := range objects {
		m := o.MinSize()
		if m.Width > w {
			w = m.Width
		}
		h += m.Height
	}
	return fyne.NewSize(w, h)
}

func (g *adaptiveGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	cols := 1
	if len(objects) > 1 && size.Width >= g.threshold {
		cols = 2
	}
	rows := (len(objects) + cols - 1) / cols
	colW := size.Width / float32(cols)

	rowH := make([]float32, rows)
	for i, o := range objects {
		r := i / cols
		if m := o.MinSize().Height; m > rowH[r] {
			rowH[r] = m
		}
	}

	y := float32(0)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			i := r*cols + c
			if i >= len(objects) {
				break
			}
			o := objects[i]
			o.Resize(fyne.NewSize(colW, rowH[r]))
			o.Move(fyne.NewPos(float32(c)*colW, y))
		}
		y += rowH[r]
	}
}
