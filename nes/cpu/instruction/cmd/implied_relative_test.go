package cmd_test

import (
	. "github.com/smarkuck/nes/nes/cpu/instruction/cmd"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	irqProgram     = 0x7ac1
	irqProgramHigh = 0x7a
	irqProgramLow  = 0xc1

	breakMarkSize = 1

	breakStatus    = status | Break | Unused
	notBreakStatus = status &^ (Break | Unused)
)

func Test_ImpliedCommands(t *T) {
	tests := []struct {
		name   string
		cmd    Implied
		before env
		after  env
	}{
		{"BRK_BreakInterrupt", BRK,
			env{ProgramCounter: prgAddr,
				Status:   status &^ InterruptDisable,
				StackPtr: InitStackPtr,
				Memory: Memory{
					IRQVector:     irqProgramLow,
					IRQVector + 1: irqProgramHigh}},
			env{ProgramCounter: irqProgram,
				Status:   status | InterruptDisable,
				StackPtr: InitStackPtr - 3,
				Stack: Stack{
					prgAddrHigh,
					prgAddrLow + breakMarkSize,
					status &^ InterruptDisable},
				Memory: Memory{
					IRQVector:     irqProgramLow,
					IRQVector + 1: irqProgramHigh}}},

		{"CLC_ClearCarryFlag", CLC,
			env{Status: status | Carry},
			env{Status: status &^ Carry}},
		{"CLD_ClearDecimalFlag", CLD,
			env{Status: status | Decimal},
			env{Status: status &^ Decimal}},
		{"CLI_ClearInterruptDisableFlag", CLI,
			env{Status: status | InterruptDisable},
			env{Status: status &^ InterruptDisable}},
		{"CLV_ClearOverflowFlag", CLV,
			env{Status: status | Overflow},
			env{Status: status &^ Overflow}},

		{"DEX_DecrementRegisterX_Positive", DEX,
			env{RegisterX: 0x02, Status: notPosStatus},
			env{RegisterX: 0x01, Status: posStatus}},
		{"DEX_DecrementRegisterX_Zero", DEX,
			env{RegisterX: 0x01, Status: notZeroStatus},
			env{RegisterX: 0x00, Status: zeroStatus}},
		{"DEX_DecrementRegisterX_Negative", DEX,
			env{RegisterX: 0x00, Status: notNegStatus},
			env{RegisterX: 0xff, Status: negStatus}},

		{"DEY_DecrementRegisterY_Positive", DEY,
			env{RegisterY: 0x02, Status: notPosStatus},
			env{RegisterY: 0x01, Status: posStatus}},
		{"DEY_DecrementRegisterY_Zero", DEY,
			env{RegisterY: 0x01, Status: notZeroStatus},
			env{RegisterY: 0x00, Status: zeroStatus}},
		{"DEY_DecrementRegisterY_Negative", DEY,
			env{RegisterY: 0x00, Status: notNegStatus},
			env{RegisterY: 0xff, Status: negStatus}},

		{"INX_IncrementRegisterX_Positive", INX,
			env{RegisterX: 0x00, Status: notPosStatus},
			env{RegisterX: 0x01, Status: posStatus}},
		{"INX_IncrementRegisterX_Zero", INX,
			env{RegisterX: 0xff, Status: notZeroStatus},
			env{RegisterX: 0x00, Status: zeroStatus}},
		{"INX_IncrementRegisterX_Negative", INX,
			env{RegisterX: 0xfe, Status: notNegStatus},
			env{RegisterX: 0xff, Status: negStatus}},

		{"INY_IncrementRegisterY_Positive", INY,
			env{RegisterY: 0x00, Status: notPosStatus},
			env{RegisterY: 0x01, Status: posStatus}},
		{"INY_IncrementRegisterY_Zero", INY,
			env{RegisterY: 0xff, Status: notZeroStatus},
			env{RegisterY: 0x00, Status: zeroStatus}},
		{"INY_IncrementRegisterY_Negative", INY,
			env{RegisterY: 0xfe, Status: notNegStatus},
			env{RegisterY: 0xff, Status: negStatus}},

		{"NOP_NoOperation", NOP, env{}, env{}},

		{"PHA_PushAccumulatorOnStack", PHA,
			env{Accumulator: value, StackPtr: InitStackPtr},
			env{Accumulator: value, StackPtr: InitStackPtr - 1,
				Stack: Stack{value}}},

		{"PHP_PushProcessorStatusOnStack", PHP,
			env{Status: status, StackPtr: InitStackPtr},
			env{Status: status, StackPtr: InitStackPtr - 1,
				Stack: Stack{status}}},

		{"PLA_PullAccumulatorFromStack_Positive", PLA,
			env{Stack: Stack{0x01}, Accumulator: 0x00,
				Status:   notPosStatus,
				StackPtr: InitStackPtr - 1},
			env{Stack: Stack{0x01}, Accumulator: 0x01,
				Status:   posStatus,
				StackPtr: InitStackPtr}},
		{"PLA_PullAccumulatorFromStack_Zero", PLA,
			env{Stack: Stack{0x00}, Accumulator: 0x01,
				Status:   notZeroStatus,
				StackPtr: InitStackPtr - 1},
			env{Stack: Stack{0x00}, Accumulator: 0x00,
				Status:   zeroStatus,
				StackPtr: InitStackPtr}},
		{"PLA_PullAccumulatorFromStack_Negative", PLA,
			env{Stack: Stack{0xff}, Accumulator: 0x00,
				Status:   notNegStatus,
				StackPtr: InitStackPtr - 1},
			env{Stack: Stack{0xff}, Accumulator: 0xff,
				Status:   negStatus,
				StackPtr: InitStackPtr}},

		{"PLP_PullProcessorStatusFromStack", PLP,
			env{Status: 0xff, StackPtr: InitStackPtr - 1,
				Stack: Stack{notBreakStatus}},
			env{Status: breakStatus, StackPtr: InitStackPtr,
				Stack: Stack{notBreakStatus}}},

		{"RTI_ReturnFromInterrupt", RTI,
			env{ProgramCounter: 0x0000, Status: 0xff,
				StackPtr: InitStackPtr - 3,
				Stack: Stack{
					prgAddrHigh, prgAddrLow, notBreakStatus}},
			env{ProgramCounter: prgAddr, Status: breakStatus,
				StackPtr: InitStackPtr,
				Stack: Stack{
					prgAddrHigh, prgAddrLow, notBreakStatus}}},

		{"RTS_ReturnFromSubroutine", RTS,
			env{ProgramCounter: 0x0000,
				StackPtr: InitStackPtr - 2,
				Stack:    Stack{prgAddrHigh, prgAddrLow}},
			env{ProgramCounter: prgAddr + subroutineOffset,
				StackPtr: InitStackPtr,
				Stack:    Stack{prgAddrHigh, prgAddrLow}}},

		{"SEC_SetCarryFlag", SEC,
			env{Status: status &^ Carry},
			env{Status: status | Carry}},
		{"SED_SetDecimalFlag", SED,
			env{Status: status &^ Decimal},
			env{Status: status | Decimal}},
		{"SEI_SetInterruptDisableFlag", SEI,
			env{Status: status &^ InterruptDisable},
			env{Status: status | InterruptDisable}},

		{"TAX_TransferAccumulatorToRegisterX_Positive", TAX,
			env{Accumulator: 0x01, RegisterX: 0x00,
				Status: notPosStatus},
			env{Accumulator: 0x01, RegisterX: 0x01,
				Status: posStatus}},
		{"TAX_TransferAccumulatorToRegisterX_Zero", TAX,
			env{Accumulator: 0x00, RegisterX: 0x01,
				Status: notZeroStatus},
			env{Accumulator: 0x00, RegisterX: 0x00,
				Status: zeroStatus}},
		{"TAX_TransferAccumulatorToRegisterX_Negative", TAX,
			env{Accumulator: 0xff, RegisterX: 0x00,
				Status: notNegStatus},
			env{Accumulator: 0xff, RegisterX: 0xff,
				Status: negStatus}},

		{"TAY_TransferAccumulatorToRegisterY_Positive", TAY,
			env{Accumulator: 0x01, RegisterY: 0x00,
				Status: notPosStatus},
			env{Accumulator: 0x01, RegisterY: 0x01,
				Status: posStatus}},
		{"TAY_TransferAccumulatorToRegisterY_Zero", TAY,
			env{Accumulator: 0x00, RegisterY: 0x01,
				Status: notZeroStatus},
			env{Accumulator: 0x00, RegisterY: 0x00,
				Status: zeroStatus}},
		{"TAY_TransferAccumulatorToRegisterY_Negative", TAY,
			env{Accumulator: 0xff, RegisterY: 0x00,
				Status: notNegStatus},
			env{Accumulator: 0xff, RegisterY: 0xff,
				Status: negStatus}},

		{"TSX_TransferStackPointerToRegisterX_Positive", TSX,
			env{StackPtr: 0x01, RegisterX: 0x00,
				Status: notPosStatus},
			env{StackPtr: 0x01, RegisterX: 0x01,
				Status: posStatus}},
		{"TSX_TransferStackPointerToRegisterX_Zero", TSX,
			env{StackPtr: 0x00, RegisterX: 0x01,
				Status: notZeroStatus},
			env{StackPtr: 0x00, RegisterX: 0x00,
				Status: zeroStatus}},
		{"TSX_TransferStackPointerToRegisterX_Negative", TSX,
			env{StackPtr: 0xff, RegisterX: 0x00,
				Status: notNegStatus},
			env{StackPtr: 0xff, RegisterX: 0xff,
				Status: negStatus}},

		{"TXA_TransferRegisterXToAccumulator_Positive", TXA,
			env{RegisterX: 0x01, Accumulator: 0x00,
				Status: notPosStatus},
			env{RegisterX: 0x01, Accumulator: 0x01,
				Status: posStatus}},
		{"TXA_TransferRegisterXToAccumulator_Zero", TXA,
			env{RegisterX: 0x00, Accumulator: 0x01,
				Status: notZeroStatus},
			env{RegisterX: 0x00, Accumulator: 0x00,
				Status: zeroStatus}},
		{"TXA_TransferRegisterXToAccumulator_Negative", TXA,
			env{RegisterX: 0xff, Accumulator: 0x00,
				Status: notNegStatus},
			env{RegisterX: 0xff, Accumulator: 0xff,
				Status: negStatus}},

		{"TXS_TransferRegisterXToStackPointer", TXS,
			env{RegisterX: 0x01, StackPtr: 0x00},
			env{RegisterX: 0x01, StackPtr: 0x01}},

		{"TYA_TransferRegisterYToAccumulator_Positive", TYA,
			env{RegisterY: 0x01, Accumulator: 0x00,
				Status: notPosStatus},
			env{RegisterY: 0x01, Accumulator: 0x01,
				Status: posStatus}},
		{"TYA_TransferRegisterYToAccumulator_Zero", TYA,
			env{RegisterY: 0x00, Accumulator: 0x01,
				Status: notZeroStatus},
			env{RegisterY: 0x00, Accumulator: 0x00,
				Status: zeroStatus}},
		{"TYA_TransferRegisterYToAccumulator_Negative", TYA,
			env{RegisterY: 0xff, Accumulator: 0x00,
				Status: notNegStatus},
			env{RegisterY: 0xff, Accumulator: 0xff,
				Status: negStatus}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			before := test.before.toState()
			test.cmd(before)
			expectStateEq(t, before, test.after.toState())
		})
	}
}

func Test_AccumulativeCommands(t *T) {
	tests := []struct {
		name   string
		cmd    Implied
		before env
		after  env
	}{
		{"ASL_ArithmeticShiftLeft_SetPositiveCarry",
			AccumASL,
			env{Accumulator: 0b10100010,
				Status: notPosStatus &^ Carry},
			env{Accumulator: 0b01000100,
				Status: posStatus | Carry}},
		{"ASL_ArithmeticShiftLeft_SetZeroCarry",
			AccumASL,
			env{Accumulator: 0b10000000,
				Status: notZeroStatus &^ Carry},
			env{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},
		{"ASL_ArithmeticShiftLeft_SetNegativeNoCarry",
			AccumASL,
			env{Accumulator: 0b01000010,
				Status: notNegStatus | Carry},
			env{Accumulator: 0b10000100,
				Status: negStatus &^ Carry}},

		{"LSR_LogicalShiftRight_SetPositiveNoCarry",
			AccumLSR,
			env{Accumulator: 0b10001010,
				Status: notPosStatus | Carry},
			env{Accumulator: 0b01000101,
				Status: posStatus &^ Carry}},
		{"LSR_LogicalShiftRight_SetZeroCarry",
			AccumLSR,
			env{Accumulator: 0b00000001,
				Status: notZeroStatus &^ Carry},
			env{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},

		{"ROL_RotateLeft_AddCarry_SetPositiveCarry",
			AccumROL,
			env{Accumulator: 0b10100010,
				Status: notPosStatus | Carry},
			env{Accumulator: 0b01000101,
				Status: posStatus | Carry}},
		{"ROL_RotateLeft_SetZeroCarry", AccumROL,
			env{Accumulator: 0b10000000,
				Status: notZeroStatus &^ Carry},
			env{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},
		{"ROL_RotateLeft_AddCarry_SetNegativeNoCarry",
			AccumROL,
			env{Accumulator: 0b01000010,
				Status: notNegStatus | Carry},
			env{Accumulator: 0b10000101,
				Status: negStatus &^ Carry}},

		{"ROR_RotateRight_SetPositiveNoCarry",
			AccumROR,
			env{Accumulator: 0b10100010,
				Status: notPosStatus &^ Carry},
			env{Accumulator: 0b01010001,
				Status: posStatus &^ Carry}},
		{"ROR_RotateRight_SetZeroCarry", AccumROR,
			env{Accumulator: 0b00000001,
				Status: notZeroStatus &^ Carry},
			env{Accumulator: 0b00000000,
				Status: zeroStatus | Carry}},
		{"ROR_RotateRight_AddCarry_SetNegativeNoCarry",
			AccumROR,
			env{Accumulator: 0b01000010,
				Status: notNegStatus | Carry},
			env{Accumulator: 0b10100001,
				Status: negStatus &^ Carry}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			before := test.before.toState()
			test.cmd(before)
			expectBinStateEq(t, before, test.after.toState())
		})
	}
}

func Test_RelativeCommands(t *T) {
	tests := []struct {
		name  string
		cmd   Relative
		flag  byte
		value bool
	}{
		{"BPL_BranchOnPlus", BPL, Negative, false},
		{"BMI_BranchOnMinus", BMI, Negative, true},
		{"BVC_BranchOnOverflowClear", BVC, Overflow, false},
		{"BVS_BranchOnOverflowSet", BVS, Overflow, true},
		{"BCC_BranchOnCarryClear", BCC, Carry, false},
		{"BCS_BranchOnCarrySet", BCS, Carry, true},
		{"BNE_BranchOnNotEqual", BNE, Zero, false},
		{"BEQ_BranchOnEqual", BEQ, Zero, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			ExpectEq(t,
				test.cmd(status|test.flag), test.value)
			ExpectEq(t,
				test.cmd(status&^test.flag), !test.value)
		})
	}
}
