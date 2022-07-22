package cmd_test

import (
	. "github.com/smarkuck/nes/nes/cpu/instruction/cmd"
	"github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	irqProgram     = 0x7ac1
	irqProgramHigh = 0x7a
	irqProgramLow  = 0xc1

	prgAddr     = 0x80fc
	prgAddrHigh = 0x80
	prgAddrLow  = 0xfc

	status           = 0b10111001
	value            = 0x7d
	breakMarkSize    = 1
	subroutineOffset = 1

	positiveStatus    = status &^ (Zero | Negative)
	notPositiveStatus = status | Zero | Negative
	zeroStatus        = (status | Zero) &^ Negative
	notZeroStatus     = negativeStatus
	negativeStatus    = (status &^ Zero) | Negative
	notNegativeStatus = zeroStatus
	breakStatus       = status | Break | Unused
	notBreakStatus    = status &^ (Break | Unused)

	invalidBusText = "invalid bus state"
)

type State struct {
	Accumulator    byte
	RegisterX      byte
	RegisterY      byte
	Status         byte
	StackPtr       byte
	ProgramCounter uint16
	Stack
	Memory
}

func (s *State) new() *state.State {
	state := new(state.State)
	state.Accumulator = s.Accumulator
	state.RegisterX = s.RegisterX
	state.RegisterY = s.RegisterY
	state.Status = s.Status
	state.StackPtr = s.StackPtr
	state.ProgramCounter = s.ProgramCounter
	state.Bus = NewTestBusStack(s.Stack, s.Memory)
	return state
}

func expectStateEq(t *T, before, after *state.State) {
	ExpectRegistersEq(t, before, after)
	ExpectStatusEq(t, before, after.Status)
	ExpectStackPtrEq(t, before, after.StackPtr)
	ExpectProgramCounterEq(t, before, after.ProgramCounter)
	expectBusEq(t, before, after)
}

func expectBusEq(t *T, before, after *state.State) {
	ExpectDeepEq(t, before.Bus, after.Bus, invalidBusText)
}

func expectAccumulatorBinEq(t *T,
	before, after *state.State) {
	ExpectBinByteEq(t, before.Accumulator, after.Accumulator,
		InvalidAccumulatorText)
}

func Test_Commands(t *T) {
	tests := []struct {
		name   string
		cmd    Implied
		before State
		after  State
	}{
		{"BRK_SoftInterrupt", BRK,
			State{
				ProgramCounter: prgAddr,
				Status:         status &^ InterruptDisable,
				StackPtr:       InitStackPtr,
				Memory: Memory{
					IRQVector:     irqProgramLow,
					IRQVector + 1: irqProgramHigh}},
			State{
				ProgramCounter: irqProgram,
				Status:         status | InterruptDisable,
				StackPtr:       InitStackPtr - 3,
				Stack: Stack{
					prgAddrHigh,
					prgAddrLow + breakMarkSize,
					status &^ InterruptDisable},
				Memory: Memory{
					IRQVector:     irqProgramLow,
					IRQVector + 1: irqProgramHigh}}},

		{"CLC_ClearCarryFlag", CLC,
			State{Status: status | Carry},
			State{Status: status &^ Carry}},
		{"CLD_ClearDecimalFlag", CLD,
			State{Status: status | Decimal},
			State{Status: status &^ Decimal}},
		{"CLI_ClearInterruptDisableFlag", CLI,
			State{Status: status | InterruptDisable},
			State{Status: status &^ InterruptDisable}},
		{"CLV_ClearOverflowFlag", CLV,
			State{Status: status | Overflow},
			State{Status: status &^ Overflow}},

		{"DEX_DecrementRegisterX_Positive", DEX,
			State{RegisterX: 0x02, Status: notPositiveStatus},
			State{RegisterX: 0x01, Status: positiveStatus}},
		{"DEX_DecrementRegisterX_Zero", DEX,
			State{RegisterX: 0x01, Status: notZeroStatus},
			State{RegisterX: 0x00, Status: zeroStatus}},
		{"DEX_DecrementRegisterX_Negative", DEX,
			State{RegisterX: 0x00, Status: notNegativeStatus},
			State{RegisterX: 0xff, Status: negativeStatus}}, // -1

		{"DEY_DecrementRegisterY_Positive", DEY,
			State{RegisterY: 0x02, Status: notPositiveStatus},
			State{RegisterY: 0x01, Status: positiveStatus}},
		{"DEY_DecrementRegisterY_Zero", DEY,
			State{RegisterY: 0x01, Status: notZeroStatus},
			State{RegisterY: 0x00, Status: zeroStatus}},
		{"DEY_DecrementRegisterY_Negative", DEY,
			State{RegisterY: 0x00, Status: notNegativeStatus},
			State{RegisterY: 0xff, Status: negativeStatus}}, // -1

		{"INX_IncrementRegisterX_Positive", INX,
			State{RegisterX: 0x00, Status: notPositiveStatus},
			State{RegisterX: 0x01, Status: positiveStatus}},
		{"INX_IncrementRegisterX_Zero", INX,
			State{RegisterX: 0xff, Status: notZeroStatus}, // -1
			State{RegisterX: 0x00, Status: zeroStatus}},
		{"INX_IncrementRegisterX_Negative", INX,
			State{RegisterX: 0xfe, Status: notNegativeStatus}, // -2
			State{RegisterX: 0xff, Status: negativeStatus}},

		{"INY_IncrementRegisterY_Positive", INY,
			State{RegisterY: 0x00, Status: notPositiveStatus},
			State{RegisterY: 0x01, Status: positiveStatus}},
		{"INY_IncrementRegisterY_Zero", INY,
			State{RegisterY: 0xff, Status: notZeroStatus}, // -1
			State{RegisterY: 0x00, Status: zeroStatus}},
		{"INY_IncrementRegisterY_Negative", INY,
			State{RegisterY: 0xfe, Status: notNegativeStatus}, // -2
			State{RegisterY: 0xff, Status: negativeStatus}},

		{"NOP_NoOperation", NOP, State{}, State{}},

		{"PHA_PushAccumulatorOnStack", PHA,
			State{Accumulator: value, StackPtr: InitStackPtr},
			State{Accumulator: value, StackPtr: InitStackPtr - 1,
				Stack: Stack{value}}},

		{"PHP_PushStatusOnStack", PHP,
			State{Status: status, StackPtr: InitStackPtr},
			State{Status: status, StackPtr: InitStackPtr - 1,
				Stack: Stack{status}}},

		{"PLA_PullAccumulatorFromStack_Positive", PLA,
			State{Stack: Stack{0x01}, Accumulator: 0x00,
				Status:   notPositiveStatus,
				StackPtr: InitStackPtr - 1},
			State{Stack: Stack{0x01}, Accumulator: 0x01,
				Status:   positiveStatus,
				StackPtr: InitStackPtr}},
		{"PLA_PullAccumulatorFromStack_Zero", PLA,
			State{Stack: Stack{0x00}, Accumulator: 0x01,
				Status:   notZeroStatus,
				StackPtr: InitStackPtr - 1},
			State{Stack: Stack{0x00}, Accumulator: 0x00,
				Status:   zeroStatus,
				StackPtr: InitStackPtr}},
		{"PLA_PullAccumulatorFromStack_Negative", PLA,
			State{Stack: Stack{0xff}, Accumulator: 0x00,
				Status:   notNegativeStatus,
				StackPtr: InitStackPtr - 1},
			State{Stack: Stack{0xff}, Accumulator: 0xff, // -1
				Status:   negativeStatus,
				StackPtr: InitStackPtr}},

		{"PLP_PullStatusFromStack", PLP,
			State{Stack: Stack{notBreakStatus}, Status: 0xff,
				StackPtr: InitStackPtr - 1},
			State{Stack: Stack{notBreakStatus}, Status: breakStatus,
				StackPtr: InitStackPtr}},

		{"RTI_ReturnFromInterrupt", RTI,
			State{ProgramCounter: 0x0000, Status: 0xff,
				StackPtr: InitStackPtr - 3,
				Stack: Stack{
					prgAddrHigh, prgAddrLow, notBreakStatus}},
			State{ProgramCounter: prgAddr, Status: breakStatus,
				StackPtr: InitStackPtr,
				Stack: Stack{
					prgAddrHigh, prgAddrLow, notBreakStatus}}},

		{"RTS_ReturnFromSubroutine", RTS,
			State{ProgramCounter: 0x0000,
				StackPtr: InitStackPtr - 2,
				Stack:    Stack{prgAddrHigh, prgAddrLow}},
			State{ProgramCounter: prgAddr + subroutineOffset,
				StackPtr: InitStackPtr,
				Stack:    Stack{prgAddrHigh, prgAddrLow}}},

		{"SEC_SetCarryFlag", SEC,
			State{Status: status &^ Carry},
			State{Status: status | Carry}},
		{"SED_SetDecimalFlag", SED,
			State{Status: status &^ Decimal},
			State{Status: status | Decimal}},
		{"SEI_SetInterruptDisableFlag", SEI,
			State{Status: status &^ InterruptDisable},
			State{Status: status | InterruptDisable}},

		{"TAX_TransferAccumulatorToRegisterX_Positive", TAX,
			State{Accumulator: 0x01, RegisterX: 0x00,
				Status: notPositiveStatus},
			State{Accumulator: 0x01, RegisterX: 0x01,
				Status: positiveStatus}},
		{"TAX_TransferAccumulatorToRegisterX_Zero", TAX,
			State{Accumulator: 0x00, RegisterX: 0x01,
				Status: notZeroStatus},
			State{Accumulator: 0x00, RegisterX: 0x00,
				Status: zeroStatus}},
		{"TAX_TransferAccumulatorToRegisterX_Negative", TAX,
			State{Accumulator: 0xff, RegisterX: 0x00,
				Status: notNegativeStatus},
			State{Accumulator: 0xff, RegisterX: 0xff, // -1
				Status: negativeStatus}},

		{"TAY_TransferAccumulatorToRegisterY_Positive", TAY,
			State{Accumulator: 0x01, RegisterY: 0x00,
				Status: notPositiveStatus},
			State{Accumulator: 0x01, RegisterY: 0x01,
				Status: positiveStatus}},
		{"TAY_TransferAccumulatorToRegisterY_Zero", TAY,
			State{Accumulator: 0x00, RegisterY: 0x01,
				Status: notZeroStatus},
			State{Accumulator: 0x00, RegisterY: 0x00,
				Status: zeroStatus}},
		{"TAY_TransferAccumulatorToRegisterY_Negative", TAY,
			State{Accumulator: 0xff, RegisterY: 0x00,
				Status: notNegativeStatus},
			State{Accumulator: 0xff, RegisterY: 0xff, // -1
				Status: negativeStatus}},

		{"TSX_TransferStackPointerToRegisterX_Positive", TSX,
			State{StackPtr: 0x01, RegisterX: 0x00,
				Status: notPositiveStatus},
			State{StackPtr: 0x01, RegisterX: 0x01,
				Status: positiveStatus}},
		{"TSX_TransferStackPointerToRegisterX_Zero", TSX,
			State{StackPtr: 0x00, RegisterX: 0x01,
				Status: notZeroStatus},
			State{StackPtr: 0x00, RegisterX: 0x00,
				Status: zeroStatus}},
		{"TSX_TransferStackPointerToRegisterX_Negative", TSX,
			State{StackPtr: 0xff, RegisterX: 0x00,
				Status: notNegativeStatus},
			State{StackPtr: 0xff, RegisterX: 0xff, // -1
				Status: negativeStatus}},

		{"TXA_TransferRegisterXToAccumulator_Positive", TXA,
			State{RegisterX: 0x01, Accumulator: 0x00,
				Status: notPositiveStatus},
			State{RegisterX: 0x01, Accumulator: 0x01,
				Status: positiveStatus}},
		{"TXA_TransferRegisterXToAccumulator_Zero", TXA,
			State{RegisterX: 0x00, Accumulator: 0x01,
				Status: notZeroStatus},
			State{RegisterX: 0x00, Accumulator: 0x00,
				Status: zeroStatus}},
		{"TXA_TransferRegisterXToAccumulator_Negative", TXA,
			State{RegisterX: 0xff, Accumulator: 0x00,
				Status: notNegativeStatus},
			State{RegisterX: 0xff, Accumulator: 0xff, // -1
				Status: negativeStatus}},

		{"TXS_TransferRegisterXToStackPointer", TXS,
			State{RegisterX: 0x01, StackPtr: 0x00},
			State{RegisterX: 0x01, StackPtr: 0x01}},

		{"TYA_TransferRegisterYToAccumulator_Positive", TYA,
			State{RegisterY: 0x01, Accumulator: 0x00,
				Status: notPositiveStatus},
			State{RegisterY: 0x01, Accumulator: 0x01,
				Status: positiveStatus}},
		{"TYA_TransferRegisterYToAccumulator_Zero", TYA,
			State{RegisterY: 0x00, Accumulator: 0x01,
				Status: notZeroStatus},
			State{RegisterY: 0x00, Accumulator: 0x00,
				Status: zeroStatus}},
		{"TYA_TransferRegisterYToAccumulator_Negative", TYA,
			State{RegisterY: 0xff, Accumulator: 0x00,
				Status: notNegativeStatus},
			State{RegisterY: 0xff, Accumulator: 0xff, // -1
				Status: negativeStatus}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			before := test.before.new()
			test.cmd(before)
			expectStateEq(t, before, test.after.new())
		})
	}
}

func Test_ShiftCommands(t *T) {
	tests := []struct {
		name   string
		cmd    Implied
		before state.State
		after  state.State
	}{
		{"ASL_ArithmeticShiftLeft_SetPositiveCarry", ASL,
			state.State{Accumulator: 0b10100010,
				Status: notPositiveStatus &^ Carry},
			state.State{Accumulator: 0b01000100,
				Status: positiveStatus | Carry}},
		{"ASL_ArithmeticShiftLeft_SetZeroCarry", ASL,
			state.State{Accumulator: 0b10000000,
				Status: notZeroStatus &^ Carry},
			state.State{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},
		{"ASL_ArithmeticShiftLeft_SetNegativeNoCarry", ASL,
			state.State{Accumulator: 0b01000010,
				Status: notNegativeStatus | Carry},
			state.State{Accumulator: 0b10000100,
				Status: negativeStatus &^ Carry}},

		{"LSR_LogicalShiftRight_SetPositiveNoCarry", LSR,
			state.State{Accumulator: 0b10001010,
				Status: notPositiveStatus | Carry},
			state.State{Accumulator: 0b01000101,
				Status: positiveStatus &^ Carry}},
		{"LSR_LogicalShiftRight_SetZeroCarry", LSR,
			state.State{Accumulator: 0b00000001,
				Status: notZeroStatus &^ Carry},
			state.State{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},

		{"ROL_RotateLeft_AddCarry_SetPositiveCarry", ROL,
			state.State{Accumulator: 0b10100010,
				Status: notPositiveStatus | Carry},
			state.State{Accumulator: 0b01000101,
				Status: positiveStatus | Carry}},
		{"ROL_RotateLeft_SetZeroCarry", ROL,
			state.State{Accumulator: 0b10000000,
				Status: notZeroStatus &^ Carry},
			state.State{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},
		{"ROL_RotateLeft_AddCarry_SetNegativeNoCarry", ROL,
			state.State{Accumulator: 0b01000010,
				Status: notNegativeStatus | Carry},
			state.State{Accumulator: 0b10000101,
				Status: negativeStatus &^ Carry}},

		{"ROR_RotateRight_SetPositiveNoCarry", ROR,
			state.State{Accumulator: 0b10100010,
				Status: notPositiveStatus &^ Carry},
			state.State{Accumulator: 0b01010001,
				Status: positiveStatus &^ Carry}},
		{"ROR_RotateRight_SetZeroCarry", ROR,
			state.State{Accumulator: 0b00000001,
				Status: notZeroStatus &^ Carry},
			state.State{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},
		{"ROR_RotateRight_AddCarry_SetNegativeNoCarry", ROR,
			state.State{Accumulator: 0b01000010,
				Status: notNegativeStatus | Carry},
			state.State{Accumulator: 0b10100001,
				Status: negativeStatus &^ Carry}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.cmd(&test.before)
			expectAccumulatorBinEq(t, &test.before, &test.after)
			ExpectStatusEq(t, &test.before, test.after.Status)
		})
	}
}
