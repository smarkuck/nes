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
	Decimal
	Break
	Unused
	Overflow
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
type Stack = []byte
type Memory = map[uint16]byte

func NewTestBusStack(s Stack, m Memory) TestBus {
	bus := TestBus{}
	bus.loadStack(s)
	bus.loadMemory(m)
	return bus
}

func (t TestBus) loadStack(s Stack) {
	for i, v := range s {
		t[InitStackAddr-uint16(i)] = v
	}
}

func NewTestBusResetPrg(addr uint16, p Program) TestBus {
	return NewTestBusProgram(addr, p,
		Memory{
			ResetVector:     byteutil.GetLow(addr),
			ResetVector + 1: byteutil.GetHigh(addr),
		},
	)
}

func NewTestBusProgram(
	addr uint16, p Program, m Memory) TestBus {
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
	ExpectRegistersEqf(t, actual, expected, byteutil.HexByte)
	ExpectStatusEq(t, actual, expected.Status)
	ExpectStackPtrEq(t, actual, expected.StackPtr)
	ExpectProgramCounterEq(t, actual, expected.ProgramCounter)
	ExpectBusEq(t, actual, expected.Bus)
}

func ExpectRegistersEqf(t test,
	actual, expected statePtr, format string) {
	ExpectAccumulatorEqf(t,
		actual, expected.Accumulator, format)
	ExpectRegisterXEqf(t,
		actual, expected.RegisterX, format)
	ExpectRegisterYEqf(t,
		actual, expected.RegisterY, format)
}

func ExpectAccumulatorEq(t test, s statePtr, value byte) {
	ExpectAccumulatorEqf(t, s, value, byteutil.HexByte)
}

func ExpectAccumulatorEqf(t test,
	s statePtr, value byte, format string) {
	unittest.ExpectEqf(t, s.Accumulator, value,
		format, invalidAccumulatorText)
}

func ExpectRegisterXEqf(t test,
	s statePtr, value byte, format string) {
	unittest.ExpectEqf(t, s.RegisterX, value,
		format, invalidRegisterXText)
}

func ExpectRegisterYEqf(t test,
	s statePtr, value byte, format string) {
	unittest.ExpectEqf(t, s.RegisterY, value,
		format, invalidRegisterYText)
}

func ExpectStatusEq(t test, s statePtr, value byte) {
	ExpectBinByteEq(t, s.Status, value,
		invalidStatusText)
}

func ExpectStackPtrEq(t test, s statePtr, value byte) {
	ExpectHexByteEq(t, s.StackPtr, value,
		invalidStackPtrText)
}

func ExpectProgramCounterEq(t test,
	s statePtr, value uint16) {
	ExpectTwoHexBytesEq(t, s.ProgramCounter, value,
		invalidProgramCounterText)
}

func ExpectBinByteEq(t test,
	actual, expected byte, msg ...string) {
	unittest.ExpectEqf(t, actual, expected,
		byteutil.BinByte, msg...)
}

func ExpectHexByteEq(t test,
	actual, expected byte, msg ...string) {
	unittest.ExpectEqf(t, actual, expected,
		byteutil.HexByte, msg...)
}

func ExpectTwoHexBytesEq(t test,
	actual, expected uint16, msg ...string) {
	unittest.ExpectEqf(t, actual, expected,
		byteutil.TwoHexBytes, msg...)
}

func ExpectBusEq(t test, s statePtr, value nes.Bus) {
	p1 := fmt.Sprintf("%p", s.Bus)
	p2 := fmt.Sprintf("%p", value)
	unittest.ExpectEq(t, p1, p2, invalidBusText)
}
