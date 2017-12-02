package CHIP8

import (
	"testing"
)

// Instruction 00E0: Clear the display.
func (cpu *CPU) TestClear() {

}

// Instruction 00EE: Return from a subroutine.
// The CPU sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func TestRet(t *testing.T) {

}

// Instruction 1nnn: Jump to location nnn.
// The CPU sets the program counter to nnn.
func TestJump(t *testing.T) {

}

// Instruction 2nnn: Call subroutine at nnn.
// The CPU increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func TestCall(t *testing.T) {

}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The CPU compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func TestSkipIf(t *testing.T) {

}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The CPU compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func TestSkipIfNot(t *testing.T) {

}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The CPU compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func TestSkipIfXY(t *testing.T) {

}

// Instruction 6xkk: Set Vx = kk.
// The CPU puts the value kk into register Vx.
func TestLoad(t *testing.T) {

}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func TestAdd(t *testing.T) {

}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func TestLoadXY(t *testing.T) {

}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corresponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func TestOrXY(t *testing.T) {
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corresponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func TestAndXY(t *testing.T) {

}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func TestXorXY(t *testing.T) {

}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func TestAddXY(t *testing.T) {

}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func TestSubXY(t *testing.T) {

}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func TestShiftRight(t *testing.T) {

}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func TestSubYX(t *testing.T) {

}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func TestShiftLeft(t *testing.T) {

}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func TestSkipIfNotXY(t *testing.T) {

}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func TestLoadI(t *testing.T) {

}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func TestJumpV0(t *testing.T) {

}

// Instruction Cxkk: Set Vx = random byte AND kk.
// The CPU generates a random number from 0 to 255,
// which is then ANDed with the value kk. The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
func TestRand(t *testing.T) {

}

// Instruction Dxyn: Display n-byte sprite starting at memory location I at (Vx, Vy),
// set VF = collision.
//
// The CPU reads n bytes from memory, starting at the address stored in I.
// These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
// Sprites are XORed onto the existing screen. If this causes any pixels to be erased,
// VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of it
// is outside the coordinates of the display, it wraps around to the opposite side of the screen.
// See instruction 8xy3 for more information on XOR, and section 2.4, Display,
// for more information on the Chip-8 screen and sprites.
func TestDraw(t *testing.T) {

}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func TestSkipIfKey(t *testing.T) {

}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func TestSkipIfKeyNot(t *testing.T) {

}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func TestLoadXDT(t *testing.T) {

}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func TestLoadDTX(t *testing.T) {

}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func TestLoadSTX(t *testing.T) {

}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func TestAddIX(t *testing.T) {

}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func TestLoadIX(t *testing.T) {

}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The CPU takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func TestLoadBCD(t *testing.T) {

}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The CPU copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func TestSaveV(t *testing.T) {

}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The CPU reads values from memory starting at location I into registers V0 through Vx.
func TestLoadV(t *testing.T) {

}
