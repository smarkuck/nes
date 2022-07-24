package cmd_test

import (
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	status        = 0b10111001
	posStatus     = status &^ (Zero | Negative)
	notPosStatus  = status | Zero | Negative
	zeroStatus    = (status | Zero) &^ Negative
	notZeroStatus = negStatus
	negStatus     = (status &^ Zero) | Negative
	notNegStatus  = zeroStatus

	prgAddr     = 0x80fc
	prgAddrHigh = 0x80
	prgAddrLow  = 0xfc

	cellAddr = 0xce11

	value            = 0x7d
	subroutineOffset = 1

	invalidBusText  = "invalid bus state"
	invalidCellText = "invalid addressed cell"
)

type State = state.State
type env = environment

type environment struct {
	Accumulator    byte
	RegisterX      byte
	RegisterY      byte
	Status         byte
	StackPtr       byte
	ProgramCounter uint16
	Cell           byte
	Stack
	Memory
}

func (s *environment) toState() *State {
	state := new(State)
	s.setFields(state)
	s.addCellToMemory()
	state.Bus = NewTestBusStack(s.Stack, s.Memory)
	return state
}

func (s *environment) setFields(state *State) {
	state.Accumulator = s.Accumulator
	state.RegisterX = s.RegisterX
	state.RegisterY = s.RegisterY
	state.Status = s.Status
	state.StackPtr = s.StackPtr
	state.ProgramCounter = s.ProgramCounter
}

func (s *environment) addCellToMemory() {
	if s.Memory == nil {
		s.Memory = Memory{cellAddr: s.Cell}
	} else if _, ok := s.Memory[cellAddr]; !ok {
		s.Memory[cellAddr] = s.Cell
	}
}

func expectStateEq(t *T, before, after *State) {
	ExpectRegistersEqf(t, before, after, byteutil.HexByte)
	ExpectStatusEq(t, before, after.Status)
	ExpectStackPtrEq(t, before, after.StackPtr)
	ExpectProgramCounterEq(t, before, after.ProgramCounter)
	expectCellEqf(t, before, after.Bus.Read(cellAddr),
		byteutil.HexByte)
	expectBusEq(t, before, after)
}

func expectBinStateEq(t *T, before, after *State) {
	ExpectRegistersEqf(t, before, after, byteutil.BinByte)
	ExpectStatusEq(t, before, after.Status)
	ExpectStackPtrEq(t, before, after.StackPtr)
	ExpectProgramCounterEq(t, before, after.ProgramCounter)
	expectCellEqf(t, before, after.Bus.Read(cellAddr),
		byteutil.BinByte)
	expectBusEq(t, before, after)
}

func expectCellEqf(t *T,
	s *State, value byte, format string) {
	ExpectEqf(t, s.Bus.Read(cellAddr), value,
		format, invalidCellText)
}

func expectBusEq(t *T, before, after *State) {
	ExpectDeepEq(t, before.Bus, after.Bus, invalidBusText)
}
