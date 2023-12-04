package fynegui

import (
	"distributed-sys-emulator/bus"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Declare conformance with the Component interface
var _ Component = (*ControlBar)(nil)

type ControlBar struct {
	*fyne.Container
}

func NewControlBar(eb bus.EventBus) *ControlBar {
	nodeCntEntry := widget.NewEntry()
	nodeCntEntry.OnChanged = func(s string) {
		nodeCntEntry.Text = extractWholeNumbers(s)
	}
	nodeCntEntry.OnSubmitted = func(s string) {
		nodeCnt, _ := strconv.Atoi(s)
		e := bus.Event{Type: bus.NodeCntChangeEvt, Data: nodeCnt}
		eb.Publish(e)
	}

	eb.Bind(bus.NetworkResizeEvt, func(resizeData bus.NetworkResize) {
		nodeCnt := resizeData.Cnt
		nodeCntEntry.Text = strconv.Itoa(nodeCnt)
		nodeCntEntry.Refresh()
	})

	var startButton, stopButton, debugButton, continueButton *widget.Button

	startButton = widget.NewButton("Start", func() {
		e := bus.Event{Type: bus.StartNodesEvt, Data: nil}
		eb.Publish(e)

		debugButton.Disable()
		startButton.Disable()
		continueButton.Disable()
		stopButton.Enable()
	})

	stopButton = widget.NewButton("Stop", func() {
		e := bus.Event{Type: bus.StopNodesEvt, Data: nil}
		eb.Publish(e)

		stopButton.Disable()
		debugButton.Disable()
		debugButton.Enable()
		startButton.Enable()
	})

	debugButton = widget.NewButton("Debug", func() {
		e := bus.Event{Type: bus.DebugNodesEvt, Data: nil}
		eb.Publish(e)

		startButton.Disable()
		debugButton.Disable()
		stopButton.Enable()
		continueButton.Enable()
	})

	continueButton = widget.NewButton("Continue", func() {
		e := bus.Event{Type: bus.ContinueNodesEvt, Data: nil}
		eb.Publish(e)

		stopButton.Disable()
		debugButton.Disable()
		startButton.Disable()
		debugButton.Disable()
	})

	continueButton.Disable()
	stopButton.Disable()

	execution := container.NewHBox(
		startButton,
		stopButton,
		widget.NewSeparator(),
		debugButton,
		continueButton,
		widget.NewSeparator(),
		nodeCntEntry,
	)

	return &ControlBar{execution}
}

func extractWholeNumbers(input string) string {
	re := regexp.MustCompile(`[0-9]+`)

	matches := re.FindAllString(input, -1)

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
