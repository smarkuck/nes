package cmd

import "github.com/smarkuck/nes/nes/cpu/state"

const (
	breakMarkSize    = 1
	subroutineOffset = 1
)

type Implied = func(*State)
type Addressed = func(_ *State, addr uint16)
type Relative = func(status byte) bool

type State = state.State

func BRK(s *State) {
	s.ProgramCounter += breakMarkSize
	s.PushTwoBytesOnStack(s.ProgramCounter)
	PHP(s)
	SEI(s)
	s.LoadIRQProgram()
}

func CLC(s *State) {
	s.DisableFlags(state.Carry)
}

func CLI(s *State) {
	s.DisableFlags(state.InterruptDisable)
}

func PHA(s *State) {
	s.PushOnStack(s.Accumulator)
}

func PHP(s *State) {
	s.PushOnStack(s.Status)
}

func PLA(s *State) {
	s.Accumulator = s.PullFromStack()
	s.UpdateZeroNegativeFlags(s.Accumulator)
}

func PLP(s *State) {
	s.Status = s.PullFromStack()
	s.EnableFlags(state.Break | state.Unused)
}

func RTI(s *State) {
	PLP(s)
	s.ProgramCounter = s.PullTwoBytesFromStack()
}

func RTS(s *State) {
	s.ProgramCounter =
		s.PullTwoBytesFromStack() + subroutineOffset
}

func SEC(s *State) {
	s.EnableFlags(state.Carry)
}

func SEI(s *State) {
	s.EnableFlags(state.InterruptDisable)
}
