package state_test

import (
	"github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

type State = state.State

const (
	address     = 0xcafe
	value16     = 0x2f9c
	value16High = 0x2f
	value16Low  = 0x9c
	value       = 0xc7
)

func Test_GetParamAddress(t *T) {
	s := State{ProgramCounter: address}

	ExpectTwoHexBytesEq(t, s.GetParamAddress(), address+1)
}

func Test_Read(t *T) {
	s := State{Bus: TestBus{address: value}}

	ExpectHexByteEq(t, s.Read(address), value)
}

func Test_ReadTwoBytes(t *T) {
	s := State{Bus: TestBus{
		0xcaff: value16Low,
		0xcb00: value16High,
	}}

	ExpectTwoHexBytesEq(t, s.ReadTwoBytes(0xcaff), value16)
}

func Test_ReadTwoBytesPageOverflow(t *T) {
	s := State{Bus: TestBus{
		0xcaff: value16Low,
		0xca00: value16High,
	}}

	ExpectTwoHexBytesEq(t,
		s.ReadTwoBytesPageOverflow(0xcaff), value16)
}

func Test_ReadInstructionCode(t *T) {
	s := State{
		ProgramCounter: address,
		Bus:            TestBus{address: value},
	}

	ExpectHexByteEq(t, s.ReadInstructionCode(), value)
}

func Test_ReadOneByteParam(t *T) {
	s := State{
		ProgramCounter: address,
		Bus:            TestBus{address + 1: value},
	}

	ExpectHexByteEq(t, s.ReadOneByteParam(), value)
}

func Test_ReadTwoBytesParam(t *T) {
	s := State{
		ProgramCounter: address,
		Bus: TestBus{
			address + 1: value16Low,
			address + 2: value16High,
		},
	}

	ExpectTwoHexBytesEq(t, s.ReadTwoBytesParam(), value16)
}

func Test_LoadResetProgram(t *T) {
	s := State{Bus: TestBus{
		ResetVector:     value16Low,
		ResetVector + 1: value16High,
	}}

	s.LoadResetProgram()

	ExpectTwoHexBytesEq(t, s.ProgramCounter, value16)
}

func Test_LoadIRQProgram(t *T) {
	s := State{Bus: TestBus{
		IRQVector:     value16Low,
		IRQVector + 1: value16High,
	}}

	s.LoadIRQProgram()

	ExpectTwoHexBytesEq(t, s.ProgramCounter, value16)
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

	ExpectHexByteEq(t, bus[address], value)
}

func Test_PushOnStack(t *T) {
	bus := TestBus{}
	s := State{StackPtr: InitStackPtr, Bus: bus}

	s.PushOnStack(value)

	ExpectHexByteEq(t, bus[InitStackAddr], value)
	ExpectHexByteEq(t, s.StackPtr, InitStackPtr-1)
}

func Test_PushTwoBytesOnStack(t *T) {
	bus := TestBus{}
	s := State{StackPtr: InitStackPtr, Bus: bus}

	s.PushTwoBytesOnStack(value16)

	ExpectHexByteEq(t, bus[InitStackAddr], value16High)
	ExpectHexByteEq(t, bus[InitStackAddr-1], value16Low)
	ExpectHexByteEq(t, s.StackPtr, InitStackPtr-2)
}

func Test_PullFromStack(t *T) {
	s := State{
		StackPtr: InitStackPtr - 1,
		Bus:      TestBus{InitStackAddr: value},
	}

	ExpectHexByteEq(t, s.PullFromStack(), value)
	ExpectHexByteEq(t, s.StackPtr, InitStackPtr)
}

func Test_PullTwoBytesFromStack(t *T) {
	s := State{
		StackPtr: InitStackPtr - 2,
		Bus: TestBus{
			InitStackAddr:     value16High,
			InitStackAddr - 1: value16Low,
		},
	}

	ExpectTwoHexBytesEq(t, s.PullTwoBytesFromStack(), value16)
	ExpectHexByteEq(t, s.StackPtr, InitStackPtr)
}

func Test_GetCarryValue(t *T) {
	s := State{Status: value &^ Carry}
	ExpectEq(t, s.GetCarryValue(), 0)

	s = State{Status: value | Carry}
	ExpectEq(t, s.GetCarryValue(), 1)
}

func Test_Flags(t *T) {
	tests := []struct {
		name string
		cmd  func(byte) bool
		flag byte
	}{
		{"IsCarry", state.IsCarry, Carry},
		{"IsZero", state.IsZero, Zero},
		{"IsOverflow", state.IsOverflow, Overflow},
		{"IsNegative", state.IsNegative, Negative},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			status := value | test.flag
			ExpectTrue(t, test.cmd(status))

			status = value &^ test.flag
			ExpectFalse(t, test.cmd(status))
		})
	}
}

func Test_EnableFlags(t *T) {
	s := State{Status: Carry | Zero}
	s.EnableFlags(Zero | Negative)
	ExpectBinByteEq(t, s.Status, Carry|Zero|Negative)
}

func Test_DisableFlags(t *T) {
	s := State{Status: Carry | Zero}
	s.DisableFlags(Zero | Negative)
	ExpectBinByteEq(t, s.Status, Carry)
}

func Test_UpdateZeroNegativeFlags(t *T) {
	tests := []struct {
		name         string
		value        byte
		statusBefore byte
		statusAfter  byte
	}{
		{"Zero", 0x00, Negative, Zero},
		{"MinPositive", 0x01, Zero | Negative, 0},
		{"MaxPositive", 0x7f, Zero | Negative, 0}, // 127
		{"MaxNegative", 0x80, Zero, Negative},     // -128
		{"MinNegative", 0xff, Zero, Negative},     // -1
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			s := State{Status: test.statusBefore}
			s.UpdateZeroNegativeFlags(test.value)
			ExpectBinByteEq(t, s.Status, test.statusAfter)
		})
	}
}

func Test_UpdateRightShiftCarryFlag(t *T) {
	s := State{Status: Carry | Zero}
	s.UpdateRightShiftCarryFlag(value &^ 0b00000001)
	ExpectBinByteEq(t, s.Status, Zero)

	s = State{Status: Zero}
	s.UpdateRightShiftCarryFlag(value | 0b00000001)
	ExpectBinByteEq(t, s.Status, Zero|Carry)
}

func Test_UpdateLeftShiftCarryFlag(t *T) {
	s := State{Status: Carry | Zero}
	s.UpdateLeftShiftCarryFlag(value &^ 0b10000000)
	ExpectBinByteEq(t, s.Status, Zero)

	s = State{Status: Zero}
	s.UpdateLeftShiftCarryFlag(value | 0b10000000)
	ExpectBinByteEq(t, s.Status, Zero|Carry)
}
