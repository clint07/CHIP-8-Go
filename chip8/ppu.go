package CHIP8

import (
	"github.com/veandco/go-sdl2/sdl"
)

type PPU struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	keypad map[sdl.Scancode]byte
}

const (
	title  = "CHIP-8"
	height = 320
	width  = 640
)

func (ppu *PPU) Init() error {
	ppu.keypad = map[sdl.Scancode]byte {
		sdl.SCANCODE_1: 0x1,
		sdl.SCANCODE_2: 0x2,
		sdl.SCANCODE_3: 0x3,
		sdl.SCANCODE_4: 0x4,
		sdl.SCANCODE_5: 0x5,
		sdl.SCANCODE_6: 0x6,
		sdl.SCANCODE_7: 0x7,
		sdl.SCANCODE_8: 0x8,
		sdl.SCANCODE_9: 0x9,
		sdl.SCANCODE_0: 0x0,
		sdl.SCANCODE_A: 0xA,
		sdl.SCANCODE_B: 0xB,
		sdl.SCANCODE_C: 0xC,
		sdl.SCANCODE_D: 0xD,
		sdl.SCANCODE_E: 0xE,
		sdl.SCANCODE_F: 0xF}

	var err error

	err = sdl.Init(sdl.INIT_VIDEO)

	if ppu.window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_SHOWN); err != nil {
		return err
	}

	if ppu.renderer, err = sdl.CreateRenderer(ppu.window, 1, 0); err != nil {
		return err
	}

	ppu.renderer.SetScale(10, 10)
	rect := sdl.Rect{0, 0, width, height}
	ppu.renderer.SetDrawColor(0, 0, 0, 1)
	ppu.renderer.FillRect(&rect)
	ppu.renderer.Present()
	return nil
}

func (ppu *PPU) destroy() {
	ppu.renderer.Destroy()
	ppu.window.Destroy()
	sdl.Quit()
}

func (ppu *PPU) Draw(gfx *[32][64]byte) {
	for i := 0; i < 32; i++ {
		for j := 0; j < 64; j++ {
			pixel := gfx[i][j]

			if pixel == 0 {
				ppu.renderer.SetDrawColor(0, 0, 0, 1)
			} else {
				ppu.renderer.SetDrawColor(255, 255, 255, 1)
			}

			ppu.renderer.DrawPoint(j, i)
		}
	}

	ppu.renderer.Present()
}

func (ppu *PPU) Poll(key *[16]bool) bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch eventType := event.(type) {
		case *sdl.QuitEvent:
			return true

		case *sdl.KeyUpEvent:
			if unpressed, ok := ppu.keypad[eventType.Keysym.Scancode]; ok {
				key[unpressed] = false
			}

		case *sdl.KeyDownEvent:
			if pressed, ok := ppu.keypad[eventType.Keysym.Scancode]; ok {
				key[pressed] = true
			}
		}

	}

	return false
}