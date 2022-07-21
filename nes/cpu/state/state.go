package state

import (
	"github.com/smarkuck/nes/nes"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
)

const (
	Carry = 1 << iota
	Zero
	InterruptDisable
	_
	Break
	Unused
	_
	Negative

	resetVector  = 0xfffc
	irqVector    = 0xfffe
	initStatus   = InterruptDisable | Break | Unused
	stackOffset  = 0x0100
	initStackPtr = 0xfd
	paramOffset  = 1
)

type State struct {
	Accumulator    byte
	RegisterX      byte
	RegisterY      byte
	Status         byte
	StackPtr       byte
	ProgramCounter uint16
	nes.Bus
}

func (s *State) Reset() {
	s.resetState()
	s.LoadResetProgram()
}

func (s *State) resetState() {
	*s = State{
		Status:   initStatus,
		StackPtr: initStackPtr,
		Bus:      s.Bus,
	}
}

func (s *State) LoadResetProgram() {
	s.ProgramCounter = s.ReadTwoBytes(resetVector)
}

func (s *State) LoadIRQProgram() {
	s.ProgramCounter = s.ReadTwoBytes(irqVector)
}

func (s *State) ReadTwoBytesParam() uint16 {
	return s.ReadTwoBytes(s.GetParamAddress())
}

func (s *State) ReadOneByteParam() byte {
	return s.Read(s.GetParamAddress())
}

func (s *State) ReadInstructionCode() byte {
	return s.Read(s.ProgramCounter)
}

func (s *State) ReadTwoBytes(addr uint16) uint16 {
	lo := s.Read(addr)
	hi := s.Read(addr + 1)
	return byteutil.Merge(hi, lo)
}

func (s *State) ReadTwoBytesPageOverflow(
	addr uint16) uint16 {
	lo := s.Read(addr)
	hi := s.Read(byteutil.IncrementLowByte(addr))
	return byteutil.Merge(hi, lo)
}

func (s *State) GetParamAddress() uint16 {
	return s.ProgramCounter + paramOffset
}

func (s *State) PushTwoBytesOnStack(value uint16) {
	s.PushOnStack(byteutil.GetHigh(value))
	s.PushOnStack(byteutil.GetLow(value))
}

func (s *State) PushOnStack(value byte) {
	s.Write(s.getStackAddr(), value)
	s.StackPtr--
}

func (s *State) PullTwoBytesFromStack() uint16 {
	lo := s.PullFromStack()
	hi := s.PullFromStack()
	return byteutil.Merge(hi, lo)
}

func (s *State) PullFromStack() byte {
	s.StackPtr++
	return s.Read(s.getStackAddr())
}

func (s *State) getStackAddr() uint16 {
	return stackOffset | uint16(s.StackPtr)
}

func (s *State) UpdateZeroNegativeFlags(value byte) {
	s.updateZeroFlag(value)
	s.updateNegativeFlag(value)
}

func (s *State) updateZeroFlag(value byte) {
	s.updateFlags(Zero, value == 0)
}

func (s *State) updateNegativeFlag(value byte) {
	s.updateFlags(Negative, byteutil.IsNegative(value))
}

func (s *State) updateFlags(flags byte, active bool) {
	if active {
		s.EnableFlags(flags)
	} else {
		s.DisableFlags(flags)
	}
}

func (s *State) EnableFlags(flags byte) {
	s.Status |= flags
}

func (s *State) DisableFlags(flags byte) {
	s.Status &^= flags
}
