package state

import (
	"github.com/smarkuck/nes/nes"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
)

const (
	Carry = 1 << iota
	Zero
	InterruptDisable
	Decimal
	Break
	Unused
	Overflow
	Negative

	initStatus = InterruptDisable | Break | Unused

	resetVector  = 0xfffc
	irqVector    = 0xfffe
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
	hi := s.Read(byteutil.IncrementLow(addr))
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

func (s *State) UpdateArithmeticFlags(
	a, b byte, sum uint16) {
	sumLo := byteutil.GetLow(sum)
	s.UpdateZeroNegative(sumLo)
	s.updateOverflow(a, b, sumLo)
	s.updateCarry(sum)
}

func (s *State) UpdateZeroNegative(value byte) {
	s.UpdateZero(value)
	s.UpdateNegative(value)
}

func (s *State) UpdateZero(value byte) {
	s.UpdateFlags(Zero, value == 0)
}

func (s *State) UpdateNegative(value byte) {
	s.UpdateFlags(Negative, byteutil.IsNegative(value))
}

func (s *State) updateOverflow(a, b, result byte) {
	isSameSign := !byteutil.IsLeftmostBit(a ^ b)
	resultSignDiffers := byteutil.IsLeftmostBit(a ^ result)
	s.UpdateFlags(Overflow, isSameSign && resultSignDiffers)
}

func (s *State) updateCarry(value uint16) {
	s.UpdateFlags(Carry, value > 0xff)
}

func (s *State) UpdateLeftShiftCarry(value byte) {
	s.UpdateFlags(Carry, byteutil.IsLeftmostBit(value))
}

func (s *State) UpdateRightShiftCarry(value byte) {
	s.UpdateFlags(Carry, byteutil.IsRightmostBit(value))
}

func (s *State) UpdateFlags(flags byte, active bool) {
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

func (s *State) GetCarry() uint8 {
	return s.Status & Carry
}

func IsCarry(status byte) bool {
	return isFlag(status, Carry)
}

func IsZero(status byte) bool {
	return isFlag(status, Zero)
}

func IsOverflow(status byte) bool {
	return isFlag(status, Overflow)
}

func IsNegative(status byte) bool {
	return isFlag(status, Negative)
}

func isFlag(status byte, flag byte) bool {
	return status&flag != 0
}
