package CHIP8

import (
	"github.com/veandco/go-sdl2/sdl"
)

type PPU struct {
	window   *sdl.Window
	renderer *sdl.Renderer
}

const (
	title  = "CHIP-8"
	height = 320
	width  = 640
)

func (self *PPU) Init() error {
	var err error

	err = sdl.Init(sdl.INIT_VIDEO)

	if self.window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_SHOWN); err != nil {
		return err
	}

	if self.renderer, err = sdl.CreateRenderer(self.window, 1, 0); err != nil {
		return err
	}

	rect := sdl.Rect{0, 0, width, height}
	self.renderer.SetDrawColor(0, 0, 0, 1)
	self.renderer.FillRect(&rect)
	self.renderer.Present()
	return nil
}

func (self *PPU) destroy() {
	self.renderer.Destroy()
	self.window.Destroy()
	sdl.Quit()
}

func (self *PPU) Draw(gfx *[32][64]byte) {

}
