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

func (self *PPU) Init() error {
	self.keypad = map[sdl.Scancode]byte {
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

	if self.window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_SHOWN); err != nil {
		return err
	}

	if self.renderer, err = sdl.CreateRenderer(self.window, 1, 0); err != nil {
		return err
	}

	self.renderer.SetScale(10, 10)
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
	for i := 0; i < 32; i++ {
		for j := 0; j < 64; j++ {
			pixel := gfx[i][j]

			if pixel == 0 {
				self.renderer.SetDrawColor(0, 0, 0, 1)
			} else {
				self.renderer.SetDrawColor(255, 255, 255, 1)
			}

			self.renderer.DrawPoint(j, i)
		}
	}

	self.renderer.Present()
}

func (self *PPU) Poll(key *[16]byte) bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch eventType := event.(type) {
		case *sdl.QuitEvent:
			return true

		case *sdl.KeyUpEvent:
			if unpressed, ok := self.keypad[eventType.Keysym.Scancode]; ok {
				//fmt.Printf("Unpressed %X\n", unpressed)
				key[unpressed] = 0
			}

		case *sdl.KeyDownEvent:
			if pressed, ok := self.keypad[eventType.Keysym.Scancode]; ok {
				//fmt.Printf("Pressed %X\n", pressed)
				key[pressed] = 1
			}
		}

	}

	return false
}