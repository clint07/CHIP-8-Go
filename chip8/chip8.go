package main

import (
	"fmt"
	"io/ioutil"
)

type CHIP8 struct {
	// CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	// Because the first 512 bytes are reserved for the interpreter, CHIP-8 programs will start at address 0x200 (512).
	RAM [4096]byte

	// CHIP-8 screen is 64x32 pixels.
	GFX [64 * 32]byte

	// 16 16-bit stack used for saving addresses before subroutines.
	Stack [16]uint

	// 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.
	V [16]byte

	// 16-bit Program counter. All programs start at 0x200.
	PC uint

	// 8-bit Stack pointer
	SP uint

	// Address register
	I uint

	// Delay timer
	DT int64

	// Sound timer
	ST int64

	// Keys
	Key [16]uint

	// ROM Size: length of CHIP-8 program byte array
	RS int

	// Draw Flag
	DF uint
}

func (chip8 *CHIP8) LoadROM(filename *string) error {
	// Read file into byte array
	rom, err := ioutil.ReadFile(*filename)
	if err != nil {
		return err
	}

	// Save ROM size
	chip8.RS = len(rom)

	// Copy program byte array into RAM
	for i, b := range rom {
		chip8.RAM[uint(i)+chip8.PC] = b
	}

	return nil
}

// Helpful for debugging
func (chip8 *CHIP8) printRAM() {
	for i := 0; i < chip8.RS+512; i++ {
		fmt.Printf("%d: %X\n", i, chip8.RAM[i])
	}
}

func (chip8 *CHIP8) getOpCode(PC uint) uint16 {
	// Each opcode is 2 bytes, but RAM is a byte array, so it must be accessed twice to create the opcode.
	//
	// Ex.
	// RAM[PC] = 0x01 (1 byte)
	// RAM[PC + 1] = 0xFE (1 byte)
	// opcode = RAM[PC] + RAM[PC + 1] = 0x01FE
	opCode1 := uint16(chip8.RAM[PC])
	opCode2 := uint16(chip8.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	fmt.Printf("\n1st OpCode: %X\n2nd OpCode: %X\nOpCode: %X\n", opCode1, opCode2, opCode)
	return opCode
}

func (chip8 *CHIP8) Cycle() {
	// Get opcode and increment PC twice
	opCode := chip8.getOpCode(chip8.PC)
	chip8.PC += 2

	// Execute code
	chip8.execute(opCode)
}


func (chip8 *CHIP8) execute(opCode uint16) {
	nnn := 	opCode & 0x0FFF
	vx := 	opCode & 0x0F00
	vy := 	opCode & 0x00F0
	kk := 	opCode & 0x00FF
	n := 	opCode & 0x000F

	if opCode == 0x00E0 {
		fmt.Println("Instruction 00E0: Clear the display.")
		chip8.clear()


	} else if opCode == 0x00EE {
		fmt.Println("Instruction 00EE: Return from a subroutine.")
		chip8.ret()


	} else if (opCode & 0xF000) == 0x1000 {
		fmt.Println("Instruction 1nnn: Jump to location nnn.")
		chip8.jmp(nnn)


	} else if (opCode & 0xF000) == 0x2000 {
		fmt.Println("Instruction 2nnn: Call subroutine at nnn.")
		chip8.call(nnn)


	} else if (opCode & 0xF000) == 0x3000 {
		fmt.Println("Instruction 3xkk: Skip next instructionif Vx = kk.")
		chip8.skipIf(vx, kk)


	} else if (opCode & 0xF000) == 0x4000 {
		fmt.Println("Instruction 4xkk: Skip next instruction if Vx != kk.")
		chip8.skipIfNot(vx, kk)


	} else if (opCode & 0xF000) == 0x5000 {
		fmt.Println("Instruction 5xy0: Skip next isntruction if Vx = Vy.")
		chip8.skipIfXY(vx, vy)


	} else if (opCode & 0xF000) == 0x6000 {
		fmt.Println("Instruction 6xkk: Set Vx = kk.")
		chip8.load(vx, kk)


	} else if (opCode & 0xF000) == 0x7000 {
		fmt.Println("Instruction 7xkk: Set Vx = Vx + kk.")
		chip8.add(vx, kk)


	} else if (opCode & 0xF00F) == 0x8000 {
		fmt.Println("Instruction 8xy0: Set Vx = Vy.")
		chip8.loadXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8001 {
		fmt.Println("Instruction 8xy1: Set Vx = Vx | Vy.")
		chip8.orXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8002 {
		fmt.Println("Instruction 8xy2: Set Vx = Vx & Vy.")
		chip8.andXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8003 {
		fmt.Println("Instruction 8xy3: Set Vx = Vx ^ Vy.")
		chip8.xorXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8004 {
		fmt.Println("Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.")
		chip8.addXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8005 {
		fmt.Println("Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.")
		chip8.subXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8006 {
		fmt.Println("Instruction 8xy6: Set Vx = Vx SHR 1.")
		chip8.shiftRight(vx)


	} else if (opCode & 0xF00F) == 0x8007 {
		fmt.Println("Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.")
		chip8.subnXY(vx, vy)


	} else if (opCode & 0xF00F) == 0x8000 {
		fmt.Println("Instruction 8xyE: Set Vx = Vx SHL 1.")
		chip8.shiftLeft(vx)


	} else if (opCode & 0xF00F) == 0x9000 {
		fmt.Println("Instruction 9xy0: Skip next instruction if Vx != Vy.")
		chip8.skipIfNotXY(vx, vy)


	} else if (opCode & 0xF000) == 0xA000 {
		fmt.Println("Instruction Annn: Set I = nnnn.")
		chip8.loadI(nnn)


	} else if (opCode & 0xF000) == 0xB000 {
		fmt.Println("Instruction Bnnn: Jump to location nnn + V0.")
		chip8.jumpV0(nnn)


	} else if (opCode & 0xF000) == 0xC000 {
		fmt.Println("Instruction Cxkk: Set Vx = random byte AND kk.")
		chip8.rand(vx)


	} else if (opCode & 0xF000) == 0xD000 {
		fmt.Println("Instruction Dxyn: Display nbyte sprite starting at memory location I at (Vx, Vy), set Vf = collusion.")
		chip8.draw(vx, vy, n)


	} else if (opCode & 0xF0FF) == 0xE09E {
		fmt.Println("Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.")
		chip8.skipIfKey(vx)


	} else if (opCode & 0xF0FF) == 0xE0A1 {
		fmt.Println("Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.")
		chip8.skipIfKeyNot(vx)


	} else if (opCode & 0xF0FF) == 0xF007 {
		fmt.Println("Instruction Fx07: Set Vx = delay timer value.")
		chip8.loadXDT(vx)


	} else if (opCode & 0xF0FF) == 0xF00A {
		fmt.Println("Instruction Fx0A: Wait for a key press, store the value of the key in Vx.")
		//chip8.loadKey(vx, key)


	} else if (opCode & 0xF0FF) == 0xF015 {
		fmt.Println("Instruction Fx15: Set delay timer = Vx.")
		chip8.loadDTX(vx)


	} else if (opCode & 0xF0FF) == 0xF018 {
		fmt.Println("Instruction Fx18: Set sounder timer = Vx.")
		chip8.loadSTX(vx)


	} else if (opCode & 0xF0FF) == 0xF01E {
		fmt.Println("Instruction Fx1E : Set I = I + Vx.")
		chip8.addIX(vx)


	} else if (opCode & 0xF0FF) == 0xF029 {
		fmt.Println("Instruction Fx29: Set I = location of sprite for digit Vx.")
		chip8.loadIX(vx)


	} else if (opCode & 0xF0FF) == 0xF033 {
		fmt.Println("Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.")
		chip8.loadBCD(vx)


	} else if (opCode & 0xF0FF) == 0xF055 {
		fmt.Println("Instruction Fx55: Store registers V0 through Vx in memory starting at location I.")
		chip8.saveV(vx)


	} else if (opCode & 0xF0FF) == 0xF065 {
		fmt.Println("Instruction Fx65: Read registers V0 through Vx in memory starting at location I.")
		chip8.loadV(vx)


	} else {
		fmt.Printf("Unknown instruction: %X\n\n", opCode)
	}

}
