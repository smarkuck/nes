package state_test

import (
	"github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

type State = state.State

const (
	pos = 0
	neg = 0x80

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
	bus := NewTestBusResetPrg(address, nil)
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

func Test_GetCarry(t *T) {
	s := State{Status: value &^ Carry}
	ExpectEq(t, s.GetCarry(), 0)

	s = State{Status: value | Carry}
	ExpectEq(t, s.GetCarry(), 1)
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

func Test_UpdateFlags(t *T) {
	s := State{Status: Carry | Zero}
	s.UpdateFlags(Zero|Negative, true)
	ExpectBinByteEq(t, s.Status, Carry|Zero|Negative)

	s = State{Status: Carry | Zero}
	s.DisableFlags(Zero | Negative)
	ExpectBinByteEq(t, s.Status, Carry)
}

func Test_UpdateZero(t *T) {
	s := State{Status: 0}
	s.UpdateZero(0)
	ExpectBinByteEq(t, s.Status, Zero)

	s = State{Status: Zero}
	s.UpdateZero(1)
	ExpectBinByteEq(t, s.Status, 0)

	s = State{Status: Zero}
	s.UpdateZero(0xff)
	ExpectBinByteEq(t, s.Status, 0)
}

func Test_UpdateNegative(t *T) {
	s := State{Status: Negative}
	s.UpdateNegative(0)
	ExpectBinByteEq(t, s.Status, 0)

	s = State{Status: Negative}
	s.UpdateNegative(0x7f)
	ExpectBinByteEq(t, s.Status, 0)

	s = State{Status: 0}
	s.UpdateNegative(0x80)
	ExpectBinByteEq(t, s.Status, Negative)

	s = State{Status: 0}
	s.UpdateNegative(0xff)
	ExpectBinByteEq(t, s.Status, Negative)
}

func Test_UpdateZeroNegative(t *T) {
	tests := []struct {
		name         string
		value        byte
		statusBefore byte
		statusAfter  byte
	}{
		{"Zero", 0, Negative, Zero},
		{"MinPositive", 1, Zero | Negative, 0},
		{"MaxPositive", 0x7f, Zero | Negative, 0},
		{"MaxNegative", 0x80, Zero, Negative},
		{"MinNegative", 0xff, Zero, Negative},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			s := State{Status: test.statusBefore}
			s.UpdateZeroNegative(test.value)
			ExpectBinByteEq(t, s.Status, test.statusAfter)
		})
	}
}

func Test_UpdateRightShiftCarry(t *T) {
	s := State{Status: Carry | Zero}
	s.UpdateRightShiftCarry(value &^ 0b00000001)
	ExpectBinByteEq(t, s.Status, Zero)

	s = State{Status: Zero}
	s.UpdateRightShiftCarry(value | 0b00000001)
	ExpectBinByteEq(t, s.Status, Zero|Carry)
}

func Test_UpdateLeftShiftCarry(t *T) {
	s := State{Status: Carry | Zero}
	s.UpdateLeftShiftCarry(value &^ 0b10000000)
	ExpectBinByteEq(t, s.Status, Zero)

	s = State{Status: Zero}
	s.UpdateLeftShiftCarry(value | 0b10000000)
	ExpectBinByteEq(t, s.Status, Zero|Carry)
}

func Test_UpdateArithmeticFlags(t *T) {
	tests := []struct {
		name         string
		a, b         byte
		sum          uint16
		statusBefore byte
		statusAfter  byte
	}{
		{"Zero", 0, 0, 0, Negative, Zero},
		{"MinPositive", 0, 1, 1, Zero | Negative, 0},
		{"MaxPositive", 0, 0x7f, 0x7f, Zero | Negative, 0},
		{"MaxNegative", 0, 0x80, 0x80, Zero, Negative},
		{"MinNegative", 0, 0xff, 0xff, Zero, Negative},
		{"Carry", 0, 0, 0x100, 0, Zero | Carry},
		{"NoOverflow", pos, pos, pos, Overflow, Zero},
		{"NoOverflow2", neg, neg, neg, Overflow, Negative},
		{"NoOverflow3", pos, neg, pos, Overflow, Zero},
		{"NoOverflow4", pos, neg, neg, Overflow, Negative},
		{"NoOverflow5", neg, pos, pos, Overflow, Zero},
		{"NoOverflow6", neg, pos, neg, Overflow, Negative},
		{"PosOverflow", pos, pos, neg, 0, Negative | Overflow},
		{"NegOverflow", neg, neg, pos, 0, Zero | Overflow},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			s := State{Status: test.statusBefore}
			s.UpdateArithmeticFlags(test.a, test.b, test.sum)
			ExpectBinByteEq(t, s.Status, test.statusAfter)
		})
	}
}
