package gui

import (
	"distributed-sys-emulator/bus"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ControlBar struct {
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*ControlBar)(nil)

func NewControlBar(eb *bus.EventBus) ControlBar {
	nodeCntEntry := widget.NewEntry()
	nodeCntEntry.OnChanged = func(s string) {
		nodeCntEntry.Text = extractWholeNumbers(s)
	}
	nodeCntEntry.OnSubmitted = func(s string) {
		nodeCnt, _ := strconv.Atoi(s)
		e := bus.Event{Type: bus.GUINodeCntChangeEvt, Data: nodeCnt}
		eb.Publish(e)
	}

	eb.Bind(bus.NetworkNodeCntChangeEvt, func(e bus.Event) {
		nodeCnt := e.Data.(int)
		nodeCntEntry.Text = strconv.Itoa(nodeCnt)
		nodeCntEntry.Refresh()
	})

	execution := container.NewHBox(
		widget.NewButton("Start", func() {
			e := bus.Event{Type: bus.StartEvt, Data: nil}
			eb.Publish(e)
		}),
		widget.NewButton("Stop", func() {
			e := bus.Event{Type: bus.StopEvt, Data: nil}
			eb.Publish(e)
		}),
		nodeCntEntry,
	)

	return ControlBar{execution}
}

func extractWholeNumbers(input string) string {
	// Define a regular expression to match whole numbers
	re := regexp.MustCompile(`[0-9]+`)

	// Find all matches in the input string
	matches := re.FindAllString(input, -1)

	// Join the matches into a single string
	result := ""
	if len(matches) > 0 {
		result = matches[0]
		for i := 1; i < len(matches); i++ {
			result += matches[i]
		}
	}

	return result
}

func (c ControlBar) GetCanvasObj() fyne.CanvasObject {
	return c.Container
}
