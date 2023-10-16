package gui

import (
	"distributed-sys-emulator/backend"
	"distributed-sys-emulator/bus"
	"encoding/json"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Canvas struct {
	// *canvas.Raster
	*fyne.Container
}

// Declare conformance with the Component interface
var _ Component = (*Canvas)(nil)

func NewCanvas(eb *bus.EventBus, wcanvas fyne.Canvas, nodeCnt int) Canvas {
	var canvasRaster *canvas.Raster
	points := pointsOnCircle(point{50, 50}, 35, nodeCnt)

	canvasc := container.NewMax()
	buttonsc := container.NewWithoutLayout()

	buttons := make([]*widget.Button, len(points))
	nodePopups := make([]*widget.PopUp, len(points))

	buttonGen := func(i int) func() {
		return func() {
			p := nodePopups[i]
			p.Resize(fyne.NewSize(300, 300))
			p.Show()
		}
	}

	for i := range buttons {
		// init popup
		label := widget.NewLabel("Custom json data to use in lua : ")
		errorLabel := widget.NewLabel("")
		errorLabel.Hide()

		jsonInput := widget.NewMultiLineEntry()
		jsonInput.PlaceHolder = `{"foo":"bar"}`
		jsonInput.Resize(fyne.NewSize(300, 300))

		jsonInput.OnChanged = func(s string) {
			// unmarshal string/check json format validity
			var data interface{}
			err := json.Unmarshal([]byte(s), &data)
			if err != nil {
				errorLabel.SetText(err.Error())
				errorLabel.Show()
				return
			}

			changeData := bus.NodeDataChangeData{TargetId: i, Data: data}
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
		nodePopups[i] = popup

		// init buttons
		nodeName := "Node"
		buttons[i] = widget.NewButton(nodeName, buttonGen(i))
		buttonsc.Add(buttons[i])
		buttons[i].Resize(buttons[i].MinSize())
	}

	// keep connections up to date
	var connections [][]*backend.Connection
	eb.Bind(bus.ConnectionChangeEvt, func(e bus.Event) {
		newConnections, ok := e.Data.([][]*backend.Connection)
		if ok {
			connections = newConnections
			canvasRaster.Refresh()
		}
	})

	canvasRaster = canvas.NewRaster(func(w, h int) image.Image {
		ratiow := float64(w) / 100
		ratioh := float64(h) / 100

		// move node popup buttons to node positions
		for i, p := range points {
			x := float32(ratiow*p.X) * .5
			y := float32(ratioh*p.Y) * .5

			popupButton := buttons[i]
			popupButton.Move(fyne.NewPos(x+2, y+2))
		}

		// draw nodes and edges
		return draw(w, h, points, &connections)
	})

	canvasc.Add(canvasRaster)
	wrap := container.NewMax(canvasc, buttonsc)
	return Canvas{wrap}
}

func (c Canvas) GetCanvasObj() fyne.CanvasObject {
	return c.Container
}

// gets width and height for the image, a network of nodes and the points that
// represent the nodes on the image but for 100x100 units
func draw(w, h int, points []point, connections *[][]*backend.Connection) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// TODO : do this once in the beginning/on every resize ?
	for x := 0; x <= w; x++ {
		for y := 0; y <= h; y++ {
			img.Set(x, y, color.White)
		}
	}

	ratiow := float64(w) / 100
	ratioh := float64(h) / 100
	lines := connLines(connections, points, ratiow, ratioh)

	// draw nodes
	for _, p := range points {
		x := int(ratiow * p.X)
		y := int(ratioh * p.Y)

		nodeRad := 4
		for nx := x - nodeRad; nx <= x+nodeRad; nx++ {
			for ny := y - nodeRad; ny <= y+nodeRad; ny++ {
				img.Set(nx, ny, color.Black)
			}
		}
	}

	// draw lines
	for _, line := range lines {
		angle := .0
		switch {
		// horizontal line
		case math.Abs(float64(line.Ymin-line.Ymax)) <= 1.0:
			for x := line.Xmin; x <= line.Xmax; x++ {
				img.Set(x, line.Ymin, color.Black)
			}

			angle = 0 // to left
			if int(line.Target.X) == line.Xmax {
				angle = math.Pi // to right
			}

		// vertical line
		case math.Abs(float64(line.Xmin-line.Xmax)) <= 1.0:
			for y := line.Ymin; y <= line.Ymax; y++ {
				img.Set(line.Xmin, y, color.Black)
			}

			angle = math.Pi / 2 // up
			if int(line.Target.Y) == line.Ymax {
				angle *= -1 // down
			}

		// regular line
		default:
			for x := int(line.Xmin); x <= int(line.Xmax); x++ {
				y := int(line.M*float64(x)) + int(line.N)
				img.Set(x, y, color.Black)
			}

			angle = math.Atan2(line.Target.Y-line.N, line.Target.X-line.M)

			// flip arrowhead
			if int(line.Target.X) > line.Xmin {
				angle = angle + math.Pi
			}
		}

		drawArrowhead(*img, line, angle)
	}

	return img
}

type point struct {
	X float64
	Y float64
}

// define a circle by it's center and radius and get n points equally spaced on
// the circles outline
func pointsOnCircle(center point, radius float64, n int) []point {
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

// a line can be described by y=mx+n and has a max and min bound for x values
// TODO : store angle and source, target instead, would improve the code
type line struct {
	M      float64
	N      float64
	Xmax   int
	Xmin   int
	Ymax   int
	Ymin   int
	Target point
}

// nodes and points are implicitly linked by their slice length, so nodes[0] is
// represented by points[0]
func connLines(connections *[][]*backend.Connection, points []point, ratiow, ratioh float64) []line {
	// iterate over the incoming edges for each node and calculate the connection
	// line
	lines := []line{}

	for sourceId, connList := range *connections {
		for _, conn := range connList {
			pTo := points[conn.Target]
			pFrom := points[sourceId]

			pTo = point{pTo.X * ratiow, pTo.Y * ratioh}
			pFrom = point{pFrom.X * ratiow, pFrom.Y * ratioh}

			xmax := int(math.Max(pFrom.X, pTo.X))
			xmin := int(math.Min(pFrom.X, pTo.X))

			ymax := int(math.Max(pFrom.Y, pTo.Y))
			ymin := int(math.Min(pFrom.Y, pTo.Y))

			m := (pTo.Y - pFrom.Y) / (pTo.X - pFrom.X)
			n := pTo.Y - (m * pTo.X)

			line := line{m, n, xmax, xmin, ymax, ymin, pTo}
			lines = append(lines, line)
		}
	}
	return lines
}

func drawArrowhead(img image.RGBA, line line, angle float64) {
	arrowSize := 40

	x1 := int(line.Target.X + float64(arrowSize)*math.Cos(angle+math.Pi/6))
	y1 := int(line.Target.Y + float64(arrowSize)*math.Sin(angle+math.Pi/6))
	x2 := int(line.Target.X + float64(arrowSize)*math.Cos(angle-math.Pi/6))
	y2 := int(line.Target.Y + float64(arrowSize)*math.Sin(angle-math.Pi/6))

	arrowheadColor := color.RGBA{0, 0, 0, 255}

	drawLineDDA(img, int(line.Target.X), int(line.Target.Y), x1, y1, arrowheadColor)
	drawLineDDA(img, int(line.Target.X), int(line.Target.Y), x2, y2, arrowheadColor)
}

func drawLineDDA(img image.RGBA, x1, y1, x2, y2 int, col color.RGBA) {
	dx := x2 - x1
	dy := y2 - y1
	steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))

	xIncrement := float64(dx) / float64(steps)
	yIncrement := float64(dy) / float64(steps)

	x := float64(x1)
	y := float64(y1)

	for i := 0; i <= steps; i++ {
		img.Set(int(x), int(y), col)
		x += xIncrement
		y += yIncrement
	}
}
