package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// TopWidget is a widget displaying the current CPU/MEM usage as a bar.
type TopWidget struct {
	*BaseWidget

	mode        string
	color       color.Color
	fillColor   color.Color
	borderColor color.Color

	history     int64
	history_arr [255]float64
	history_ptr int
}

// NewTopWidget returns a new TopWidget.
func NewTopWidget(bw *BaseWidget, opts WidgetConfig) *TopWidget {
	bw.setInterval(time.Duration(opts.Interval)*time.Millisecond, time.Second/2)

	var mode string
	_ = ConfigValue(opts.Config["mode"], &mode)
	var color, fillColor, borderColor color.Color
	var history int64
	_ = ConfigValue(opts.Config["color"], &color)
	_ = ConfigValue(opts.Config["fillColor"], &fillColor)
	_ = ConfigValue(opts.Config["borderColor"], &borderColor)
	_ = ConfigValue(opts.Config["history"], &history)

	return &TopWidget{
		BaseWidget:  bw,
		mode:        mode,
		color:       color,
		fillColor:   fillColor,
		borderColor: borderColor,
		history:     history,
	}
}

// Update renders the widget.
func (w *TopWidget) Update() error {
	var value, bar_width, bar_height float64
	var label string
	var x0, x1, y0, y1 int

	switch w.mode {
	case "cpu":
		cpuUsage, err := cpu.Percent(0, false)
		if err != nil {
			return fmt.Errorf("can't retrieve CPU usage: %s", err)
		}

		value = cpuUsage[0]
		label = "CPU"

	case "memory":
		memory, err := mem.VirtualMemory()
		if err != nil {
			return fmt.Errorf("can't retrieve memory usage: %s", err)
		}
		value = memory.UsedPercent
		label = "MEM"

	default:
		return fmt.Errorf("unknown widget mode: %s", w.mode)
	}

	w.history_arr[w.history_ptr] = value

	if w.color == nil {
		w.color = DefaultColor
	}
	if w.fillColor == nil {
		w.fillColor = color.RGBA{166, 155, 182, 255}
	}
	if w.borderColor == nil {
		w.borderColor = DefaultColor
	}

	size := int(w.dev.Pixels)
	margin := size / 18
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	x0 = 12
	y0 = 6
	x1 = size-x0
	y1 = size-18

	draw.Draw(img, image.Rect(x0, y0, x1, y1),
		&image.Uniform{w.borderColor},
		image.Point{}, draw.Src)

	x0 += 1
	y0 += 1
	x1 -= 1
	y1 -= 1

	draw.Draw(img, image.Rect(x0, y0, x1, y1),
		&image.Uniform{color.RGBA{0, 0, 0, 255}},
		image.Point{}, draw.Src)

	//verbosef("history %d %d %v", w.history, w.history_ptr, w.history_arr)

	bar_width = float64(x1-x0)/float64(w.history)

	for i := 0; i < int(w.history); i++ {
		index := (i + w.history_ptr - int(w.history) + len(w.history_arr)) % len(w.history_arr)

		bar_height = (100-w.history_arr[index])*float64(y1-y0)/100

		//verbosef("%d x0 %d x1 %d  y0 %d y1 %d index %d bar_height %d history_arr %f", i, x0, x1, y0, y1, index, bar_height, w.history_arr[index])
		draw.Draw(img,
			image.Rect(x0+int(bar_width*float64(i)), y0+int(bar_height), x0+int(bar_width*float64(i+1)), y1),
			&image.Uniform{w.fillColor},
			image.Point{}, draw.Src)
	}

	w.history_ptr = (w.history_ptr + 1) % len(w.history_arr)
	//verbosef("---")

	// draw percentage
	bounds := img.Bounds()
	bounds.Min.Y = 6
	bounds.Max.Y -= 18

	drawString(img,
		bounds,
		ttfFont,
		strconv.FormatInt(int64(value), 10),
		w.dev.DPI,
		13,
		w.color,
		image.Pt(-1, -1))

	// draw description
	bounds = img.Bounds()
	bounds.Min.Y = size - 16
	bounds.Max.Y -= margin

	drawString(img,
		bounds,
		ttfFont,
		"% "+label,
		w.dev.DPI,
		-1,
		w.color,
		image.Pt(-1, -1))

	return w.render(w.dev, img)
}
