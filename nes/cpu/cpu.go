package cpu

import (
	"fmt"

	"github.com/smarkuck/nes/nes"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/instruction"
	"github.com/smarkuck/nes/nes/cpu/state"
)

const (
	initStatus   = state.DisableInterrupt | state.Break | state.Unused
	initStackPtr = 0xfd

	unknownInstrFormat = "unknown instruction code: " +
		byteutil.HexByte
	invalidCyclesFormat = "encountered instruction needs " +
		"0 cycles to execute: " + byteutil.HexByte
)

type Instruction = instruction.Instruction
type State = state.State

type CPU interface {
	Tick()
	Reset()
	GetAccumulator() byte
	GetRegisterX() byte
	GetRegisterY() byte
	GetStatus() byte
	GetStackPtr() byte
	GetProgramCounter() uint16
	GetRemainingCycles() uint8
}

type cpu struct {
	Instructions
	State
	remainingCycles uint8
}

type Instructions map[byte]Instruction

func NewCPU(b nes.Bus, i Instructions) CPU {
	c := new(cpu)
	c.Bus, c.Instructions = b, i
	c.Reset()
	return c
}

func (c *cpu) Reset() {
	c.resetState()
	c.remainingCycles = 0
	c.LoadResetProgram()
}

func (c *cpu) resetState() {
	c.State = State{
		Status:   initStatus,
		StackPtr: initStackPtr,
		Bus:      c.Bus,
	}
}

func (c *cpu) Tick() {
	if c.remainingCycles == 0 {
		code, instr := c.getInstruction()
		instr.Execute(&c.State)
		c.updateCycles(code, instr)
	}
	c.remainingCycles--
}

func (c *cpu) getInstruction() (byte, Instruction) {
	code := c.Read(c.ProgramCounter)
	if i, ok := c.Instructions[code]; ok {
		return code, i
	}
	panic(getUnknownInstrErr(code))
}

func getUnknownInstrErr(code byte) error {
	return fmt.Errorf(unknownInstrFormat, code)
}

func (c *cpu) updateCycles(code byte, i Instruction) {
	c.remainingCycles = i.GetCycles()
	if c.remainingCycles == 0 {
		panic(getInvalidCyclesErr(code))
	}
}

func getInvalidCyclesErr(code byte) error {
	return fmt.Errorf(invalidCyclesFormat, code)
}

func (c *cpu) GetAccumulator() byte {
	return c.Accumulator
}

func (c *cpu) GetRegisterX() byte {
	return c.RegisterX
}

func (c *cpu) GetRegisterY() byte {
	return c.RegisterY
}

func (c *cpu) GetStatus() byte {
	return c.Status
}

func (c *cpu) GetStackPtr() byte {
	return c.StackPtr
}

func (c *cpu) GetProgramCounter() uint16 {
	return c.ProgramCounter
}

func (c *cpu) GetRemainingCycles() uint8 {
	return c.remainingCycles
}
