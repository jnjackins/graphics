package main

import (
	"image"
	"os"
	"runtime"

	"code.google.com/p/jnj.plan9/draw"
	"code.google.com/p/jnj/die"
	"github.com/jnjackins/graphics/text"
)

var (
	disp   *draw.Display
	screen *draw.Image
	buf    *text.Buffer
)

type mouseEvent struct {
	m draw.Mouse
}

func (e mouseEvent) Pos() image.Point {
	return e.m.Point
}

func (e mouseEvent) Buttons() int {
	return e.m.Buttons
}

func main() {
	var err error
	runtime.GOMAXPROCS(runtime.NumCPU())
	gopath := os.Getenv("GOPATH")
	fontpath := gopath + "/src/code.google.com/p/jnj/cmd/buf/proggy.ttf"
	buf, err = text.NewBuffer(image.Rect(0, 0, 300, 300), fontpath)
	die.On(err)
	buf.LoadString("test")
	buf.Select(text.Address{0, 0}, text.Address{0, 0})
	disp, err = draw.Init(nil, "buf", "300x322")
	die.On(err)
	defer disp.Close()
	screen = disp.ScreenImage
	mouse := disp.InitMouse()
	kbd := disp.InitKeyboard()
	redraw()

loop:
	for {
		select {
		case <-mouse.Resize:
			resize()
		case me := <-mouse.C:
			buf.SendMouseEvent(me.Point, me.Buttons)
		case ke := <-kbd.C:
			// esc
			if ke == 27 {
				break loop
			}
			buf.SendKey(ke)
		}
		if buf.Dirty() {
			redraw()
		}
	}
}

func redraw() {
	_, err := screen.Load(screen.Bounds(), buf.Img().Pix)
	die.On(err)
	disp.Flush()
}

func resize() {
	err := disp.Attach(draw.Refmesg)
	die.On(err, "error reattaching display after resize")
	buf.Resize(screen.Bounds())
}
