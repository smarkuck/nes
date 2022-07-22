package cmd

import "github.com/smarkuck/nes/nes/cpu/state"

const (
	breakMarkSize    = 1
	subroutineOffset = 1
)

type Implied = func(*state.State)
type Addressed = func(_ *state.State, addr uint16)
type Relative = func(status byte) bool

func ASL(s *state.State) {
	s.UpdateLeftShiftCarryFlag(s.Accumulator)
	s.Accumulator <<= 1
	s.UpdateZeroNegativeFlags(s.Accumulator)
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

func DEX(s *state.State) {
	s.RegisterX--
	s.UpdateZeroNegativeFlags(s.RegisterX)
}

func DEY(s *state.State) {
	s.RegisterY--
	s.UpdateZeroNegativeFlags(s.RegisterY)
}

func INX(s *state.State) {
	s.RegisterX++
	s.UpdateZeroNegativeFlags(s.RegisterX)
}

func INY(s *state.State) {
	s.RegisterY++
	s.UpdateZeroNegativeFlags(s.RegisterY)
}

func LSR(s *state.State) {
	s.UpdateRightShiftCarryFlag(s.Accumulator)
	s.Accumulator >>= 1
	s.UpdateZeroNegativeFlags(s.Accumulator)
}

func NOP(s *state.State) {}

func PHA(s *state.State) {
	s.PushOnStack(s.Accumulator)
}

func PHP(s *state.State) {
	s.PushOnStack(s.Status)
}

func PLA(s *state.State) {
	s.Accumulator = s.PullFromStack()
	s.UpdateZeroNegativeFlags(s.Accumulator)
}

func PLP(s *state.State) {
	s.Status = s.PullFromStack()
	s.EnableFlags(state.Break | state.Unused)
}

func ROL(s *state.State) {
	c := s.GetCarryValue()
	s.UpdateLeftShiftCarryFlag(s.Accumulator)
	s.Accumulator = (s.Accumulator << 1) | c
	s.UpdateZeroNegativeFlags(s.Accumulator)
}

func ROR(s *state.State) {
	c := s.GetCarryValue()
	s.UpdateRightShiftCarryFlag(s.Accumulator)
	s.Accumulator = (s.Accumulator >> 1) | c<<7
	s.UpdateZeroNegativeFlags(s.Accumulator)
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

func SEC(s *state.State) {
	s.EnableFlags(state.Carry)
}

func SED(s *state.State) {
	s.EnableFlags(state.Decimal)
}

func SEI(s *state.State) {
	s.EnableFlags(state.InterruptDisable)
}

func TAX(s *state.State) {
	s.RegisterX = s.Accumulator
	s.UpdateZeroNegativeFlags(s.RegisterX)
}

func TAY(s *state.State) {
	s.RegisterY = s.Accumulator
	s.UpdateZeroNegativeFlags(s.RegisterY)
}

func TSX(s *state.State) {
	s.RegisterX = s.StackPtr
	s.UpdateZeroNegativeFlags(s.RegisterX)
}

func TXA(s *state.State) {
	s.Accumulator = s.RegisterX
	s.UpdateZeroNegativeFlags(s.Accumulator)
}

func TXS(s *state.State) {
	s.StackPtr = s.RegisterX
}

func TYA(s *state.State) {
	s.Accumulator = s.RegisterY
	s.UpdateZeroNegativeFlags(s.Accumulator)
}
