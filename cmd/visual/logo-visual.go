package main

import (
	"io"
	"math"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	"rs.lab/go-logo/logo"
)

const SCREEN_WIDTH = 640
const SCREEN_HEIGHT = 480

var COLORMAP = map[logo.Color]uint32{
	logo.Black:   0x000000ff,
	logo.White:   0xffffffff,
	logo.Red:     0xff0000ff,
	logo.Green:   0x00ff00ff,
	logo.Blue:    0x0000ffff,
	logo.Yellow:  0xffff00ff,
	logo.Gray:    0x888888ff,
	logo.Magenta: 0xff00ffff,
}

type Visual struct {
	logo.DrawingStub
	Window   *sdl.Window
	Renderer *sdl.Renderer
}

func NewVisual() *Visual {
	return &Visual{}
}

func (v *Visual) colorToRGBA(color logo.Color) (r, g, b, a uint8) {
	val, ok := COLORMAP[color]
	if !ok {
		// Fallback
		r, g, b, a = 0, 0, 0, 0
		return
	}

	r = uint8((val >> 24) & 0xff)
	g = uint8((val >> 16) & 0xff)
	b = uint8((val >> 8) & 0xff)
	a = uint8((val) & 0xff)
	return
}

func (v *Visual) DrawTurtle(r *logo.Runtime, size float64) {
	v.Renderer.SetDrawColor(v.colorToRGBA(logo.Red))

	t := r.DegToRad(r.Angle)
	// Huh, too much  math :(
	ax, ay := r.Head.X+size*math.Cos(t), r.Head.Y+size*math.Sin(t)
	px, py := r.Head.X+size/8*math.Cos(t), r.Head.Y+size/8*math.Sin(t)
	bx, by := r.Head.X+size*math.Cos(t+2*math.Pi/3), r.Head.Y+size*math.Sin(t+2*math.Pi/3)
	cx, cy := r.Head.X+size*math.Cos(t-2*math.Pi/3), r.Head.Y+size*math.Sin(t-2*math.Pi/3)

	v.Renderer.DrawLine(int32(ax), int32(ay), int32(bx), int32(by))
	v.Renderer.DrawLine(int32(ax), int32(ay), int32(cx), int32(cy))
	v.Renderer.DrawLine(int32(bx), int32(by), int32(cx), int32(cy))
	v.Renderer.DrawLine(int32(ax), int32(ay), int32(px), int32(py))
}

func (v *Visual) Clear(r *logo.Runtime) {
	v.Renderer.SetDrawColor(v.colorToRGBA(r.Paper))
	v.Renderer.Clear()
}

func (v *Visual) DrawLine(r *logo.Runtime, x1, y1, x2, y2 int32) {
	v.Renderer.SetDrawColor(v.colorToRGBA(r.Ink))
	v.Renderer.DrawLine(x1, y1, x2, y2)
}

func main() {
	source, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	sdl.LogSetPriority(sdl.LOG_CATEGORY_APPLICATION, sdl.LOG_PRIORITY_VERBOSE)

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	visual := NewVisual()

	visual.Window, err = sdl.CreateWindow("Logo | Press 'ESCAPE' to quit", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, SCREEN_WIDTH, SCREEN_HEIGHT, 0)

	if err != nil {
		panic(err)
	}
	defer visual.Window.Destroy()

	visual.Renderer, err = sdl.CreateRenderer(visual.Window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	defer visual.Renderer.Destroy()

	r := logo.NewRuntime()
	// r.Trace = true
	r.Stub = visual
	err = r.Run(string(source))

	visual.DrawTurtle(r, 10)

	visual.Renderer.SetLogicalSize(SCREEN_WIDTH, SCREEN_HEIGHT)
	visual.Renderer.Present()

	quit := false
	for !quit {
		event := sdl.WaitEvent()
		if event != nil {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				quit = true
			case *sdl.KeyboardEvent:
				if t.Keysym.Sym == sdl.K_ESCAPE {
					quit = true
				}
			}
		}
	}

	if err != nil {
		panic(err)
	}

}
