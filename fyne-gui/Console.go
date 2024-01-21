package fynegui

import (
	"distributed-sys-emulator/bus"
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Declare conformance with the Component interface
var _ Component = (*Console)(nil)

type Console struct {
	*fyne.Container
}

func NewConsole(eb bus.EventBus) *Console {
	c := container.NewBorder(nil, nil, nil, nil)

	var nodeCnt int
	var outputs []string
	var results []any

	// depending on the outputs, create console boxes
	refresh := func() {
		c.RemoveAll()

		headerRow := container.NewGridWithColumns(nodeCnt)
		for i := 0; i < nodeCnt; i++ {
			content := "Node " + strconv.Itoa(i)

			label := canvas.NewText(content, color.RGBA{51, 153, 153, 201})
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{
				Bold:      true,
				Italic:    false,
				Monospace: false,
				Symbol:    false,
				TabWidth:  2,
			}
			entry := container.NewCenter(label)

			headerRow.Add(entry)
		}

		outRow := container.NewGridWithColumns(len(outputs))
		for i := 0; i < len(outputs); i++ {
			outputLabel := widget.NewLabel(outputs[i])
			entry := container.NewScroll(outputLabel)
			entry.SetMinSize(fyne.NewSize(0, 100))
			outRow.Add(entry)
		}

		resRow := container.NewGridWithColumns(len(results))
		for i := 0; i < len(results); i++ {
			resStr := fmt.Sprintf("%v", results[i])
			entry := widget.NewLabel(resStr)
			resRow.Add(entry)
		}

		out := container.NewVBox(
			headerRow,
			widget.NewSeparator(),
			outRow,
			widget.NewSeparator(),
			resRow,
		)

		c.Add(out)
		c.Refresh()
	}

	// update node count and output slice size as required
	eb.Bind(bus.NetworkResizeEvt, func(resizeData bus.NetworkResize) {
		nodeCnt = resizeData.Cnt
		if nodeCnt > len(outputs) {
			for i := 0; i <= nodeCnt-len(outputs); i++ {
				outputs = append(outputs, "")
				results = append(results, nil)
			}
		} else {
			outputs = outputs[:nodeCnt]
			results = results[:nodeCnt]
		}
		refresh()
	})

	// update outputs
	eb.Bind(bus.NodeOutputEvt, func(out bus.NodeOutput) {
		outputs[out.NodeId] = out.Log
		results[out.NodeId] = out.Result
		refresh()
	})

	console := &Console{c}
	return console
}

func (e *Console) GetCanvasObj() fyne.CanvasObject {
	return e.Container
}
