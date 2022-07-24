package cmd

import (
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/state"
)

const (
	breakMarkSize    = 1
	subroutineOffset = 1
)

type Implied = func(*state.State)
type Addressed = func(_ *state.State, addr uint16)
type Relative = func(status byte) bool

func ADC(s *state.State, addr uint16) {
	add(s, s.Read(addr))
}

func AND(s *state.State, addr uint16) {
	s.Accumulator &= s.Read(addr)
	s.UpdateZeroNegative(s.Accumulator)
}

func ASL(s *state.State, addr uint16) {
	b := s.Read(addr)
	asl(s, &b)
	s.Write(addr, b)
}

func AccumASL(s *state.State) {
	asl(s, &s.Accumulator)
}

func asl(s *state.State, cell *byte) {
	s.UpdateLeftShiftCarry(*cell)
	*cell <<= 1
	s.UpdateZeroNegative(*cell)
}

func BCC(status byte) bool {
	return !state.IsCarry(status)
}

func BCS(status byte) bool {
	return state.IsCarry(status)
}

func BEQ(status byte) bool {
	return state.IsZero(status)
}

func BIT(s *state.State, addr uint16) {
	value := s.Read(addr)
	s.UpdateZero(s.Accumulator & value)
	s.UpdateNegative(value)
	s.UpdateFlags(state.Overflow,
		byteutil.IsBit(value, 6))
}

func BMI(status byte) bool {
	return state.IsNegative(status)
}

func BNE(status byte) bool {
	return !state.IsZero(status)
}

func BPL(status byte) bool {
	return !state.IsNegative(status)
}

func BRK(s *state.State) {
	s.ProgramCounter += breakMarkSize
	s.PushTwoBytesOnStack(s.ProgramCounter)
	s.PushOnStack(s.Status)
	s.EnableFlags(state.InterruptDisable)
	s.LoadIRQProgram()
}

func BVC(status byte) bool {
	return !state.IsOverflow(status)
}

func BVS(status byte) bool {
	return state.IsOverflow(status)
}

func CLC(s *state.State) {
	s.DisableFlags(state.Carry)
}

func CLD(s *state.State) {
	s.DisableFlags(state.Decimal)
}

func CLI(s *state.State) {
	s.DisableFlags(state.InterruptDisable)
}

func CLV(s *state.State) {
	s.DisableFlags(state.Overflow)
}

func CMP(s *state.State, addr uint16) {
	compare(s, s.Accumulator, s.Read(addr))
}

func CPX(s *state.State, addr uint16) {
	compare(s, s.RegisterX, s.Read(addr))
}

func CPY(s *state.State, addr uint16) {
	compare(s, s.RegisterY, s.Read(addr))
}

func compare(s *state.State, a, b byte) {
	s.UpdateZeroNegative(a - b)
	s.UpdateFlags(state.Carry, a >= b)
}

func DEC(s *state.State, addr uint16) {
	v := s.Read(addr) - 1
	s.Write(addr, v)
	s.UpdateZeroNegative(v)
}

func DEX(s *state.State) {
	s.RegisterX--
	s.UpdateZeroNegative(s.RegisterX)
}

func DEY(s *state.State) {
	s.RegisterY--
	s.UpdateZeroNegative(s.RegisterY)
}

func EOR(s *state.State, addr uint16) {
	s.Accumulator ^= s.Read(addr)
	s.UpdateZeroNegative(s.Accumulator)
}

func INC(s *state.State, addr uint16) {
	v := s.Read(addr) + 1
	s.Write(addr, v)
	s.UpdateZeroNegative(v)
}

func INX(s *state.State) {
	s.RegisterX++
	s.UpdateZeroNegative(s.RegisterX)
}

func INY(s *state.State) {
	s.RegisterY++
	s.UpdateZeroNegative(s.RegisterY)
}

func JMP(s *state.State, addr uint16) {
	s.ProgramCounter = addr
}

func JSR(s *state.State, addr uint16) {
	s.PushTwoBytesOnStack(
		s.ProgramCounter - subroutineOffset)
	s.ProgramCounter = addr
}

func LDA(s *state.State, addr uint16) {
	s.Accumulator = s.Read(addr)
	s.UpdateZeroNegative(s.Accumulator)
}

func LDX(s *state.State, addr uint16) {
	s.RegisterX = s.Read(addr)
	s.UpdateZeroNegative(s.RegisterX)
}

func LDY(s *state.State, addr uint16) {
	s.RegisterY = s.Read(addr)
	s.UpdateZeroNegative(s.RegisterY)
}

func LSR(s *state.State, addr uint16) {
	b := s.Read(addr)
	lsr(s, &b)
	s.Write(addr, b)
}

func AccumLSR(s *state.State) {
	lsr(s, &s.Accumulator)
}

func lsr(s *state.State, cell *byte) {
	s.UpdateRightShiftCarry(*cell)
	*cell >>= 1
	s.UpdateZeroNegative(*cell)
}

func NOP(s *state.State) {}

func ORA(s *state.State, addr uint16) {
	s.Accumulator |= s.Read(addr)
	s.UpdateZeroNegative(s.Accumulator)
}

func PHA(s *state.State) {
	s.PushOnStack(s.Accumulator)
}

func PHP(s *state.State) {
	s.PushOnStack(s.Status)
}

func PLA(s *state.State) {
	s.Accumulator = s.PullFromStack()
	s.UpdateZeroNegative(s.Accumulator)
}

func PLP(s *state.State) {
	s.Status = s.PullFromStack()
	s.EnableFlags(state.Break | state.Unused)
}

func ROL(s *state.State, addr uint16) {
	b := s.Read(addr)
	rol(s, &b)
	s.Write(addr, b)
}

func AccumROL(s *state.State) {
	rol(s, &s.Accumulator)
}

func rol(s *state.State, cell *byte) {
	carryBit := s.GetCarry()
	s.UpdateLeftShiftCarry(*cell)
	*cell = (*cell << 1) | carryBit
	s.UpdateZeroNegative(*cell)
}

func ROR(s *state.State, addr uint16) {
	b := s.Read(addr)
	ror(s, &b)
	s.Write(addr, b)
}

func AccumROR(s *state.State) {
	ror(s, &s.Accumulator)
}

func ror(s *state.State, cell *byte) {
	carryBit := s.GetCarry() << 7
	s.UpdateRightShiftCarry(*cell)
	*cell = (*cell >> 1) | carryBit
	s.UpdateZeroNegative(*cell)
}

func RTI(s *state.State) {
	s.Status = s.PullFromStack()
	s.EnableFlags(state.Break | state.Unused)
	s.ProgramCounter = s.PullTwoBytesFromStack()
}

func RTS(s *state.State) {
	s.ProgramCounter =
		s.PullTwoBytesFromStack() + subroutineOffset
}

func SBC(s *state.State, addr uint16) {
	invertedValue := ^s.Read(addr)
	add(s, invertedValue)
}

func add(s *state.State, value byte) {
	a := s.Accumulator
	sum := uint16(a) + uint16(value) + uint16(s.GetCarry())
	s.Accumulator = byteutil.GetLow(sum)
	s.UpdateArithmeticFlags(a, value, sum)
}

func SEC(s *state.State) {
	s.EnableFlags(state.Carry)
}

func SED(s *state.State) {
	s.EnableFlags(state.Decimal)
}

func SEI(s *state.State) {
	s.EnableFlags(state.InterruptDisable)
}

func STA(s *state.State, addr uint16) {
	s.Write(addr, s.Accumulator)
}

func STX(s *state.State, addr uint16) {
	s.Write(addr, s.RegisterX)
}

func STY(s *state.State, addr uint16) {
	s.Write(addr, s.RegisterY)
}

func TAX(s *state.State) {
	s.RegisterX = s.Accumulator
	s.UpdateZeroNegative(s.RegisterX)
}

func TAY(s *state.State) {
	s.RegisterY = s.Accumulator
	s.UpdateZeroNegative(s.RegisterY)
}

func TSX(s *state.State) {
	s.RegisterX = s.StackPtr
	s.UpdateZeroNegative(s.RegisterX)
}

func TXA(s *state.State) {
	s.Accumulator = s.RegisterX
	s.UpdateZeroNegative(s.Accumulator)
}

func TXS(s *state.State) {
	s.StackPtr = s.RegisterX
}

func TYA(s *state.State) {
	s.Accumulator = s.RegisterY
	s.UpdateZeroNegative(s.Accumulator)
}
