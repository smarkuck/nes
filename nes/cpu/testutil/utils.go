package testutil

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/smarkuck/nes/nes"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/state"
	"github.com/smarkuck/unittest"
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

	InitStatus = InterruptDisable | Break | Unused

	ResetVector   = 0xfffc
	IRQVector     = 0xfffe
	StackOffset   = 0x0100
	InitStackPtr  = 0xfd
	InitStackAddr = StackOffset | InitStackPtr

	invalidAccumulatorText    = "invalid accumulator"
	invalidRegisterXText      = "invalid register X"
	invalidRegisterYText      = "invalid register Y"
	invalidStatusText         = "invalid status"
	invalidStackPtrText       = "invalid stack pointer"
	invalidProgramCounterText = "invalid program counter"
	invalidBusText            = "invalid bus reference"
)

type statePtr = *state.State
type test = *testing.T

type TestBus map[uint16]byte
type Program = []byte
type Memory = map[uint16]byte

func NewTestBusResetPrg(addr uint16, code byte) TestBus {
	return NewTestBus(ResetVector,
		Program{
			byteutil.GetLow(addr),
			byteutil.GetHigh(addr),
		},
		Memory{addr: code},
	)
}

func NewTestBus(addr uint16, p Program, m Memory) TestBus {
	bus := TestBus{}
	bus.loadProgram(addr, p)
	bus.loadMemory(m)
	return bus
}

func (t TestBus) loadProgram(addr uint16, p Program) {
	for i, v := range p {
		t[addr+uint16(i)] = v
	}
}

func (t TestBus) loadMemory(m Memory) {
	for k, v := range m {
		t[k] = v
	}
}

func (t TestBus) Read(addr uint16) byte {
	return t[addr]
}

func (t TestBus) Write(addr uint16, value byte) {
	t[addr] = value
}

func (t TestBus) String() string {
	if len(t) == 0 {
		return "empty"
	}
	return t.getString()
}

func (t TestBus) getString() string {
	e := t.getEntries()
	sort.Strings(e)
	return strings.Join(e, ", ")
}

func (t TestBus) getEntries() []string {
	format := byteutil.TwoHexBytes + ": " + byteutil.HexByte
	result := []string{}
	for k, v := range t {
		entry := fmt.Sprintf(format, k, v)
		result = append(result, entry)
	}
	return result
}

func NewInitState(addr uint16, bus nes.Bus) statePtr {
	return &state.State{
		Accumulator:    0,
		RegisterX:      0,
		RegisterY:      0,
		Status:         InitStatus,
		StackPtr:       InitStackPtr,
		ProgramCounter: addr,
		Bus:            bus,
	}
}

func NewState(value byte, bus nes.Bus) statePtr {
	return &state.State{
		Accumulator:    value,
		RegisterX:      value,
		RegisterY:      value,
		Status:         value,
		StackPtr:       value,
		ProgramCounter: uint16(value),
		Bus:            bus,
	}
}

func ExpectStateEq(t test, actual, expected statePtr) {
	ExpectRegistersEq(t, actual, expected)
	ExpectStatusEq(t, actual, expected.Status)
	ExpectStackPtrEq(t, actual, expected.StackPtr)
	ExpectProgramCounterEq(t, actual, expected.ProgramCounter)
	ExpectBusEq(t, actual, expected.Bus)
}

func ExpectRegistersEq(t test, actual, expected statePtr) {
	ExpectAccumulatorEq(t, actual, expected.Accumulator)
	ExpectRegisterXEq(t, actual, expected.RegisterX)
	ExpectRegisterYEq(t, actual, expected.RegisterY)
}

func ExpectAccumulatorEq(t test, s statePtr, value byte) {
	unittest.ExpectEq(t, s.Accumulator, value,
		invalidAccumulatorText)
}

func ExpectRegisterXEq(t test, s statePtr, value byte) {
	unittest.ExpectEq(t, s.RegisterX, value,
		invalidRegisterXText)
}

func ExpectRegisterYEq(t test, s statePtr, value byte) {
	unittest.ExpectEq(t, s.RegisterY, value,
		invalidRegisterYText)
}

func ExpectStatusEq(t test, s statePtr, value byte) {
	unittest.ExpectEqf(t, s.Status, value,
		byteutil.BinByte, invalidStatusText)
}

func ExpectStackPtrEq(t test, s statePtr, value byte) {
	unittest.ExpectEqf(t, s.StackPtr, value,
		byteutil.HexByte, invalidStackPtrText)
}

func ExpectProgramCounterEq(t test,
	s statePtr, value uint16) {
	unittest.ExpectEqf(t, s.ProgramCounter, value,
		byteutil.TwoHexBytes, invalidProgramCounterText)
}

func ExpectBusEq(t test, s statePtr, value nes.Bus) {
	p1 := fmt.Sprintf("%p", s.Bus)
	p2 := fmt.Sprintf("%p", value)
	unittest.ExpectEq(t, p1, p2, invalidBusText)
}
