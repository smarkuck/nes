package cpu

import "fmt"

const (
	initStatus   = 0x34
	initStackPtr = 0xfd
	resetVector  = 0xfffc

	hexByte                 = "%#02x"
	missingBusText          = "missing CPU bus"
	missingInstructionsText = "missing CPU instructions"
	unknownInstrFormat      = "unknown instruction code: " + hexByte
	invalidCyclesFormat     = "encountered instruction needs " +
		"0 cycles to execute: " + hexByte
)

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
	State
	Instructions
	remainingCycles uint8
}

type State struct {
	Accumulator    byte
	RegisterX      byte
	RegisterY      byte
	Status         byte
	StackPtr       byte
	ProgramCounter uint16
	Bus
}

type Bus interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
}

type Instructions map[byte]Instruction

type Instruction interface {
	Execute(*State)
	GetCycles() uint8
}

func NewCPU(b Bus, i Instructions) CPU {
	c := new(cpu)
	c.Bus = b
	c.Instructions = i
	c.Reset()
	return c
}

func (c *cpu) Reset() {
	c.resetState()
	c.remainingCycles = 0
	c.loadProgram()
}

func (c *cpu) resetState() {
	c.State = State{
		Status:   initStatus,
		StackPtr: initStackPtr,
		Bus:      c.Bus,
	}
}

func (c *cpu) loadProgram() {
	lo := c.Read(resetVector)
	hi := c.Read(resetVector + 1)
	c.ProgramCounter = uint16(hi)<<8 + uint16(lo)
}

func (c *cpu) Tick() {
	if c.remainingCycles == 0 {
		code, i := c.getInstruction()
		i.Execute(&c.State)
		c.updateCycles(code, i)
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
