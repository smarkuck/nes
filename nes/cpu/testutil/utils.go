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

	InvalidAccumulatorText    = "invalid accumulator"
	InvalidRegisterXText      = "invalid register X"
	InvalidRegisterYText      = "invalid register Y"
	InvalidStatusText         = "invalid status"
	InvalidStackPtrText       = "invalid stack pointer"
	InvalidProgramCounterText = "invalid program counter"
	InvalidBusText            = "invalid bus reference"
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

func NewTestBusResetPrg(addr uint16, code byte) TestBus {
	return NewTestBusProgram(ResetVector,
		Program{
			byteutil.GetLow(addr),
			byteutil.GetHigh(addr),
		},
		Memory{addr: code},
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

func ExpectBinByteEq(t test,
	actual, expected byte, msg ...string) {
	unittest.ExpectEqf(t, actual, expected,
		byteutil.BinByte, msg...)
}

func ExpectHexByteEq(t test, actual, expected byte) {
	unittest.ExpectEqf(t, actual, expected,
		byteutil.HexByte)
}

func ExpectTwoHexBytesEq(t test, actual, expected uint16) {
	unittest.ExpectEqf(t, actual, expected,
		byteutil.TwoHexBytes)
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
	unittest.ExpectEqf(t, s.Accumulator, value,
		byteutil.HexByte, InvalidAccumulatorText)
}

func ExpectRegisterXEq(t test, s statePtr, value byte) {
	unittest.ExpectEqf(t, s.RegisterX, value,
		byteutil.HexByte, InvalidRegisterXText)
}

func ExpectRegisterYEq(t test, s statePtr, value byte) {
	unittest.ExpectEqf(t, s.RegisterY, value,
		byteutil.HexByte, InvalidRegisterYText)
}

func ExpectStatusEq(t test, s statePtr, value byte) {
	unittest.ExpectEqf(t, s.Status, value,
		byteutil.BinByte, InvalidStatusText)
}

func ExpectStackPtrEq(t test, s statePtr, value byte) {
	unittest.ExpectEqf(t, s.StackPtr, value,
		byteutil.HexByte, InvalidStackPtrText)
}

func ExpectProgramCounterEq(t test,
	s statePtr, value uint16) {
	unittest.ExpectEqf(t, s.ProgramCounter, value,
		byteutil.TwoHexBytes, InvalidProgramCounterText)
}

func ExpectBusEq(t test, s statePtr, value nes.Bus) {
	p1 := fmt.Sprintf("%p", s.Bus)
	p2 := fmt.Sprintf("%p", value)
	unittest.ExpectEq(t, p1, p2, InvalidBusText)
}
