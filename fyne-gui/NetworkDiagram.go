package fynegui

import (
	"distributed-sys-emulator/bus"
	"encoding/json"
	"math"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/x/fyne/widget/diagramwidget"
)

type NetworkDiagram struct {
	fyne.Widget
	stateMu sync.Mutex

	// state data
	buttons []*widget.Button
	nodes   []diagramwidget.DiagramNode
	edges   []edge
}

type point struct {
	X float64
	Y float64
}

type edge struct {
	From int
	To   int
	*diagramwidget.BaseDiagramLink
}

func NewNetworkDiagram(eb bus.EventBus, wcanvas fyne.Canvas) *NetworkDiagram {
	networkDiag := NetworkDiagram{}
	networkDiag.stateMu = sync.Mutex{}

	diag := diagramwidget.NewDiagramWidget("Network Diagram")

	eb.Bind(bus.SentToEvt, func(task bus.SendTask) {
		networkDiag.refreshNodeSent(task)
	})

	eb.Bind(bus.AwaitStartEvt, func(id bus.NodeId) {
		networkDiag.refreshNodeAwait(id)
	})

	eb.Bind(bus.AwaitEndEvt, func(sendTasks []bus.SendTask) {
		networkDiag.refreshTransmitted(sendTasks)
	})

	eb.Bind(bus.NetworkResizeEvt, func(resizeData bus.NetworkResize) {
		networkDiag.refreshButtons(eb, wcanvas, resizeData.Cnt)
		networkDiag.refreshNodes(diag, resizeData.Cnt)
		networkDiag.refreshConnections(diag, resizeData.Connections)
	})

	eb.Bind(bus.NetworkConnectionsEvt, func(newConnections bus.Connections) {
		networkDiag.refreshConnections(diag, newConnections)
	})

	eb.Bind(bus.ContinueNodesEvt, func() {
		networkDiag.refreshOnContinue(diag)
	})

	diag.Refresh()
	scroll := container.NewScroll(diag)
	networkDiag.Widget = scroll
	return &networkDiag
}

func (networkDiag *NetworkDiagram) refreshOnContinue(diag *diagramwidget.DiagramWidget) {
	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	// reset source and midpoint decorations by removing and recreating edges
	// TODO : add functionality to remove decorations directly to fyne-x
	for _, e := range networkDiag.edges {
		diag.RemoveElement(e.GetDiagramElementID())
		c := bus.Connection{From: e.From, To: e.To}
		networkDiag.createLink(c, diag)
	}
	networkDiag.Refresh()
}

func (networkDiag *NetworkDiagram) refreshButtons(eb bus.EventBus,
	wcanvas fyne.Canvas, nodeCnt int) {

	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	buttons := make([]*widget.Button, nodeCnt)
	nodeModals := make([]Modal, nodeCnt)

	onPress := func(i int) func() {
		return func() {
			p := nodeModals[i]
			p.Resize(fyne.NewSize(300, 300))
			p.Show()
		}
	}

	for i := range buttons {
		// init popup
		label := widget.NewLabel("Custom json data : ")
		errorLabel := widget.NewLabel("")
		errorLabel.Hide()

		jsonInput := widget.NewMultiLineEntry()
		jsonInput.PlaceHolder = `{"foo":"bar"}`
		jsonInput.Resize(fyne.NewSize(300, 300))
		jsonInput.OnChanged = func(s string) {

			// unmarshal string/check json format validity
			var data any
			err := json.Unmarshal([]byte(s), &data)
			if err != nil {
				errorLabel.SetText(err.Error())
				errorLabel.Show()
				return
			}

			changeData := bus.NodeData{TargetId: i, Data: data}
			evt := bus.Event{Type: bus.NodeDataChangeEvt, Data: changeData}
			eb.Publish(evt)
			errorLabel.Hide()
		}

		vstack := container.NewVBox(
			label,
			jsonInput,
			errorLabel)

		popup := NewModal(vstack, wcanvas)
		popup.Hide()
		nodeModals[i] = popup

		// init buttons
		nodeName := "Node " + strconv.Itoa(i)
		buttons[i] = widget.NewButton(nodeName, onPress(i))
		buttons[i].Resize(buttons[i].MinSize())
	}

	networkDiag.buttons = buttons
	networkDiag.Refresh()
}

func (networkDiag *NetworkDiagram) refreshNodes(
	diag *diagramwidget.DiagramWidget, nodeCnt int) {

	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	points := placePointsOnCircle(nodeCnt)

	// TODO : check if count changed
	for _, n := range networkDiag.nodes {
		diag.RemoveElement(n.GetDiagramElementID())
	}
	networkDiag.nodes = networkDiag.nodes[:0]

	for i, p := range points {
		// needs an inital size because its not set for the widget on the
		// initial pass
		diagWidth := diag.Size().Width
		if diagWidth == 0 {
			diagWidth = 800
		}

		diagHeight := diag.Size().Height
		if diagHeight == 0 {
			diagHeight = 1200
		}

		ratiow := diagWidth / 100
		ratioh := diagHeight / 100

		// draw node
		x := ratiow * float32(p.X*.5)
		y := ratioh * float32(p.Y*.5)

		nodeName := "Node" + strconv.Itoa(i)
		nodeButton := networkDiag.buttons[i]
		node := diagramwidget.NewDiagramNode(diag, nodeButton, "Id:"+nodeName)
		node.Move(fyne.Position{X: x, Y: y})
		networkDiag.nodes = append(networkDiag.nodes, node)
	}
	networkDiag.Refresh()
}

func (networkDiag *NetworkDiagram) refreshConnections(
	diag *diagramwidget.DiagramWidget, connections bus.Connections) {

	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	// remove existing edges
	for _, e := range networkDiag.edges {
		diag.RemoveElement(e.GetDiagramElementID())
	}
	networkDiag.edges = networkDiag.edges[:0]

	// add edges based on new connections
	for _, c := range connections {
		networkDiag.createLink(c, diag)
	}
	networkDiag.Refresh()
}

func (networkDiag *NetworkDiagram) createLink(c bus.Connection, diag *diagramwidget.DiagramWidget) {
	linkName := "Link" + strconv.Itoa(c.From) + ":" + strconv.Itoa(c.To)
	link := diagramwidget.NewDiagramLink(diag, linkName)
	link.SetSourcePad(networkDiag.nodes[c.From].GetEdgePad())
	link.SetTargetPad(networkDiag.nodes[c.To].GetEdgePad())
	link.AddTargetDecoration(diagramwidget.NewArrowhead())
	edge := edge{c.From, c.To, link}
	networkDiag.edges = append(networkDiag.edges, edge)
}

func (networkDiag *NetworkDiagram) refreshNodeSent(task bus.SendTask) {
	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	for _, e := range networkDiag.edges {
		if e.From == task.From && e.To == task.To {
			id := "Sent:" + strconv.Itoa(e.From) + ":" + strconv.Itoa(e.To)
			e.AddSourceAnchoredText(id, "sending")
		}
	}
	networkDiag.Refresh()
}

func (networkDiag *NetworkDiagram) refreshNodeAwait(id bus.NodeId) {
	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	status := widget.NewLabel("Awaiting")
	obj := container.NewHBox(status, networkDiag.buttons[id])
	networkDiag.nodes[id].SetInnerObject(obj)
	networkDiag.Refresh()
}

func (networkDiag *NetworkDiagram) refreshTransmitted(sendTasks []bus.SendTask) {
	networkDiag.stateMu.Lock()
	defer networkDiag.stateMu.Unlock()

	for _, t := range sendTasks {
		// display synced data
		for _, e := range networkDiag.edges {
			if t.From == e.From && t.To == e.To {
				dataStr, _ := json.Marshal(t.Data)
				e.AddMidpointAnchoredText("sendTask", string(dataStr))
			}
		}

		// remove awaiting status
		networkDiag.nodes[t.To].SetInnerObject(networkDiag.buttons[t.To])
	}
	networkDiag.Refresh()
}

func placePointsOnCircle(n int) []point {
	center := point{50, 50}
	radius := 35.

	var points []point
	angleIncrement := 2 * math.Pi / float64(n)
	for i := 0; i < n; i++ {
		angle := float64(i) * angleIncrement
		x := center.X + radius*math.Cos(angle)
		y := center.Y + radius*math.Sin(angle)
		points = append(points, point{x, y})
	}

	return points
}
