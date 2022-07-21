package cpu

import (
	"fmt"

	"github.com/smarkuck/nes/nes"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/instruction"
	"github.com/smarkuck/nes/nes/cpu/state"
)

const (
	unknownInstrFormat = "unknown instruction code: " +
		byteutil.HexByte

	invalidCyclesFormat = "encountered instruction needs " +
		"0 cycles to execute: " + byteutil.HexByte
)

type instr = instruction.Instruction

type CPU interface {
	Tick()
	Reset()
	GetState() *state.State
	GetRemainingCycles() uint8
}

type cpu struct {
	Instructions
	state.State
	remainingCycles uint8
}

type Instructions map[byte]instr

func NewCPU(b nes.Bus, i Instructions) CPU {
	c := new(cpu)
	c.Bus, c.Instructions = b, i
	c.Reset()
	return c
}

func (c *cpu) Reset() {
	c.State.Reset()
	c.remainingCycles = 0
}

func (c *cpu) Tick() {
	if c.remainingCycles == 0 {
		code, instr := c.getInstruction()
		instr.Execute(&c.State)
		c.updateCycles(code, instr)
	}
	c.remainingCycles--
}

func (c *cpu) getInstruction() (byte, instr) {
	code := c.ReadInstructionCode()
	if i, ok := c.Instructions[code]; ok {
		return code, i
	}
	panic(getUnknownInstrErr(code))
}

func getUnknownInstrErr(code byte) error {
	return fmt.Errorf(unknownInstrFormat, code)
}

func (c *cpu) updateCycles(code byte, i instr) {
	c.remainingCycles = i.GetCycles()
	if c.remainingCycles == 0 {
		panic(getInvalidCyclesErr(code))
	}
}

func getInvalidCyclesErr(code byte) error {
	return fmt.Errorf(invalidCyclesFormat, code)
}

func (c *cpu) GetState() *state.State {
	s := c.State
	return &s
}

func (c *cpu) GetRemainingCycles() uint8 {
	return c.remainingCycles
}
