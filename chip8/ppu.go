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

func (this *PPU) Init() error {
	var err error

	err = sdl.Init(sdl.INIT_VIDEO)

	if this.window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_SHOWN); err != nil {
		return err
	}

	if this.renderer, err = sdl.CreateRenderer(this.window, 1, 0); err != nil {
		return err
	}

	rect := sdl.Rect{0, 0, width, height}
	this.renderer.SetDrawColor(0, 0, 0, 1)
	this.renderer.FillRect(&rect)
	this.renderer.Present()
	return nil
}

func (this *PPU) destroy() {
	this.renderer.Destroy()
	this.window.Destroy()
	sdl.Quit()
}
