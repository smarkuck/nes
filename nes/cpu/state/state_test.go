package state_test

import (
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

type State = state.State

const (
	address     = 0xcafe
	value       = 0xc7
	value16     = 0x2f9c
	value16High = 0x2f
	value16Low  = 0x9c
)

func expectBinByteEq(t *T, actual, expected byte) {
	ExpectEqf(t, actual, expected, byteutil.BinByte)
}

func expectHexByteEq(t *T, actual, expected byte) {
	ExpectEqf(t, actual, expected, byteutil.HexByte)
}

func expectTwoHexBytesEq(t *T, actual, expected uint16) {
	ExpectEqf(t, actual, expected, byteutil.TwoHexBytes)
}

func Test_GetParamAddress(t *T) {
	s := State{ProgramCounter: address}

	expectTwoHexBytesEq(t, s.GetParamAddress(), address+1)
}

func Test_Read(t *T) {
	s := State{Bus: TestBus{address: value}}

	expectHexByteEq(t, s.Read(address), value)
}

func Test_ReadTwoBytes(t *T) {
	s := State{Bus: TestBus{
		0xcaff: value16Low,
		0xcb00: value16High,
	}}

	expectTwoHexBytesEq(t, s.ReadTwoBytes(0xcaff), value16)
}

func Test_ReadTwoBytesPageOverflow(t *T) {
	s := State{Bus: TestBus{
		0xcaff: value16Low,
		0xca00: value16High,
	}}

	expectTwoHexBytesEq(t,
		s.ReadTwoBytesPageOverflow(0xcaff), value16)
}

func Test_ReadInstructionCode(t *T) {
	s := State{
		ProgramCounter: address,
		Bus:            TestBus{address: value},
	}

	expectHexByteEq(t, s.ReadInstructionCode(), value)
}

func Test_ReadOneByteParam(t *T) {
	s := State{
		ProgramCounter: address,
		Bus:            TestBus{address + 1: value},
	}

	expectHexByteEq(t, s.ReadOneByteParam(), value)
}

func Test_ReadTwoBytesParam(t *T) {
	s := State{
		ProgramCounter: address,
		Bus: TestBus{
			address + 1: value16Low,
			address + 2: value16High,
		},
	}

	expectTwoHexBytesEq(t, s.ReadTwoBytesParam(), value16)
}

func Test_LoadResetProgram(t *T) {
	s := State{Bus: TestBus{
		ResetVector:     value16Low,
		ResetVector + 1: value16High,
	}}

	s.LoadResetProgram()

	expectTwoHexBytesEq(t, s.ProgramCounter, value16)
}

func Test_LoadIRQProgram(t *T) {
	s := State{Bus: TestBus{
		IRQVector:     value16Low,
		IRQVector + 1: value16High,
	}}

	s.LoadIRQProgram()

	expectTwoHexBytesEq(t, s.ProgramCounter, value16)
}

func Test_OnReset_ClearState_LoadProgram_KeepOldBus(t *T) {
	bus := NewTestBusResetPrg(address, 0x00)
	s := NewState(value, bus)

	s.Reset()

	ExpectStateEq(t, s, NewInitState(address, bus))
}

func Test_Write(t *T) {
	bus := TestBus{}
	s := State{Bus: bus}

	s.Write(address, value)

	expectHexByteEq(t, bus[address], value)
}

func Test_PushOnStack(t *T) {
	bus := TestBus{}
	s := State{StackPtr: InitStackPtr, Bus: bus}

	s.PushOnStack(value)

	expectHexByteEq(t, bus[InitStackAddr], value)
	expectHexByteEq(t, s.StackPtr, InitStackPtr-1)
}

func Test_PushTwoBytesOnStack(t *T) {
	bus := TestBus{}
	s := State{StackPtr: InitStackPtr, Bus: bus}

	s.PushTwoBytesOnStack(value16)

	expectHexByteEq(t, bus[InitStackAddr], value16High)
	expectHexByteEq(t, bus[InitStackAddr-1], value16Low)
	expectHexByteEq(t, s.StackPtr, InitStackPtr-2)
}

func Test_PullFromStack(t *T) {
	s := State{
		StackPtr: InitStackPtr - 1,
		Bus:      TestBus{InitStackAddr: value},
	}

	expectHexByteEq(t, s.PullFromStack(), value)
	expectHexByteEq(t, s.StackPtr, InitStackPtr)
}

func Test_PullTwoBytesFromStack(t *T) {
	s := State{
		StackPtr: InitStackPtr - 2,
		Bus: TestBus{
			InitStackAddr:     value16High,
			InitStackAddr - 1: value16Low,
		},
	}

	expectTwoHexBytesEq(t, s.PullTwoBytesFromStack(), value16)
	expectHexByteEq(t, s.StackPtr, InitStackPtr)
}

func Test_EnableFlags(t *T) {
	s := State{Status: Carry | Zero}
	s.EnableFlags(Zero | Negative)
	expectBinByteEq(t, s.Status, Carry|Zero|Negative)
}

func Test_DisableFlags(t *T) {
	s := State{Status: Carry | Zero}
	s.DisableFlags(Zero | Negative)
	expectBinByteEq(t, s.Status, Carry)
}

func Test_UpdateZeroNegativeFlags(t *T) {
	tests := []struct {
		name         string
		value        byte
		statusBefore byte
		statusAfter  byte
	}{
		{"MinPositive", 0x01, Zero | Negative, 0},
		{"MaxPositive", 0x7f, Zero | Negative, 0}, // 127
		{"Zero", 0x00, Negative, Zero},
		{"MinNegative", 0xff, Zero, Negative}, // -1
		{"MaxNegative", 0x80, Zero, Negative}, // -128
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			s := State{Status: test.statusBefore}
			s.UpdateZeroNegativeFlags(test.value)
			expectBinByteEq(t, s.Status, test.statusAfter)
		})
	}
}
