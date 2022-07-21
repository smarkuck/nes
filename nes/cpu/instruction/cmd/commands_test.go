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

	prgAddr     = 0x80fc
	prgAddrHigh = 0x80
	prgAddrLow  = 0xfc

	status           = 0b10111001
	breakMarkSize    = 1
	subroutineOffset = 1

	invalidBusText = "invalid bus state"
)

func expectStateEq(t *T, before, after *State) {
	ExpectRegistersEq(t, before, after)
	ExpectStatusEq(t, before, after.Status)
	ExpectStackPtrEq(t, before, after.StackPtr)
	ExpectProgramCounterEq(t, before, after.ProgramCounter)
	expectBusEq(t, before, after)
}

func expectBusEq(t *T, before, after *State) {
	ExpectDeepEq(t, before.Bus, after.Bus, invalidBusText)
}

func Test_ImpliedInstructions(t *T) {
	tests := []struct {
		name   string
		cmd    Implied
		before State
		after  State
	}{
		{"BRK_SoftInterrupt", BRK,
			State{
				Status:         status &^ InterruptDisable,
				StackPtr:       InitStackPtr,
				ProgramCounter: prgAddr,
				Bus: TestBus{
					IRQVector:     irqProgramLow,
					IRQVector + 1: irqProgramHigh},
			},
			State{
				Status:         status | InterruptDisable,
				StackPtr:       InitStackPtr - 3,
				ProgramCounter: irqProgram,
				Bus: TestBus{
					InitStackAddr:     prgAddrHigh,
					InitStackAddr - 1: prgAddrLow + breakMarkSize,
					InitStackAddr - 2: status &^ InterruptDisable,
					IRQVector:         irqProgramLow,
					IRQVector + 1:     irqProgramHigh},
			},
		},

		{"CLI_ClearInterruptDisableFlag", CLI,
			State{Status: status | InterruptDisable},
			State{Status: status &^ InterruptDisable},
		},

		{"CLC_ClearCarryFlag", CLC,
			State{Status: status | Carry},
			State{Status: status &^ Carry},
		},

		{"PHA_PushAccumulatorOnStack", PHA,
			State{
				Accumulator: 125,
				StackPtr:    InitStackPtr,
				Bus:         TestBus{},
			},
			State{
				Accumulator: 125,
				StackPtr:    InitStackPtr - 1,
				Bus:         TestBus{InitStackAddr: 0x7d}, // 0x7d -> 125
			},
		},

		{"PHP_PushStatusOnStack", PHP,
			State{
				Status:   status,
				StackPtr: InitStackPtr,
				Bus:      TestBus{},
			},
			State{
				Status:   status,
				StackPtr: InitStackPtr - 1,
				Bus:      TestBus{InitStackAddr: status},
			},
		},

		{"PLA_PullAccumulatorFromStack_MinPositive", PLA,
			State{
				Accumulator: 0,
				Status:      status | Zero | Negative,
				StackPtr:    InitStackPtr - 1,
				Bus:         TestBus{InitStackAddr: 0x01},
			},
			State{
				Accumulator: 1,
				Status:      status &^ (Zero | Negative),
				StackPtr:    InitStackPtr,
				Bus:         TestBus{InitStackAddr: 0x01},
			},
		},

		{"PLA_PullAccumulatorFromStack_MaxPositive", PLA,
			State{
				Accumulator: 0,
				Status:      status | Zero | Negative,
				StackPtr:    InitStackPtr - 1,
				Bus:         TestBus{InitStackAddr: 0x7f}, // 0x7f -> 127
			},
			State{
				Accumulator: 127,
				Status:      status &^ (Zero | Negative),
				StackPtr:    InitStackPtr,
				Bus:         TestBus{InitStackAddr: 0x7f},
			},
		},

		{"PLA_PullAccumulatorFromStack_Zero", PLA,
			State{
				Accumulator: 1,
				Status:      (status &^ Zero) | Negative,
				StackPtr:    InitStackPtr - 1,
				Bus:         TestBus{InitStackAddr: 0x00},
			},
			State{
				Accumulator: 0,
				Status:      (status | Zero) &^ Negative,
				StackPtr:    InitStackPtr,
				Bus:         TestBus{InitStackAddr: 0x00},
			},
		},

		{"PLA_PullAccumulatorFromStack_MinNegative", PLA,
			State{
				Accumulator: 0,
				Status:      (status | Zero) &^ Negative,
				StackPtr:    InitStackPtr - 1,
				Bus:         TestBus{InitStackAddr: 0xff},
			},
			State{
				Accumulator: 0xff, // -1
				Status:      (status &^ Zero) | Negative,
				StackPtr:    InitStackPtr,
				Bus:         TestBus{InitStackAddr: 0xff},
			},
		},

		{"PLA_PullAccumulatorFromStack_MaxNegative", PLA,
			State{
				Accumulator: 0,
				Status:      (status | Zero) &^ Negative,
				StackPtr:    InitStackPtr - 1,
				Bus:         TestBus{InitStackAddr: 0x80},
			},
			State{
				Accumulator: 0x80, // -128
				Status:      (status &^ Zero) | Negative,
				StackPtr:    InitStackPtr,
				Bus:         TestBus{InitStackAddr: 0x80},
			},
		},

		{"PLP_PopRegisterFromStack", PLP,
			State{
				Status:   0b11111111,
				StackPtr: InitStackPtr - 1,
				Bus: TestBus{
					InitStackAddr: status &^ (Break | Unused)},
			},
			State{
				Status:   status | Break | Unused,
				StackPtr: InitStackPtr,
				Bus: TestBus{
					InitStackAddr: status &^ (Break | Unused)},
			},
		},

		{"RTI_ReturnFromInterrupt", RTI,
			State{
				Status:         0b11111111,
				StackPtr:       InitStackPtr - 3,
				ProgramCounter: 0x0000,
				Bus: TestBus{
					InitStackAddr:     prgAddrHigh,
					InitStackAddr - 1: prgAddrLow,
					InitStackAddr - 2: status &^ (Break | Unused)},
			},
			State{
				Status:         status | Break | Unused,
				StackPtr:       InitStackPtr,
				ProgramCounter: prgAddr,
				Bus: TestBus{
					InitStackAddr:     prgAddrHigh,
					InitStackAddr - 1: prgAddrLow,
					InitStackAddr - 2: status &^ (Break | Unused)},
			},
		},

		{"RTS_ReturnFromSubroutine", RTS,
			State{
				StackPtr:       InitStackPtr - 2,
				ProgramCounter: 0x00,
				Bus: TestBus{
					InitStackAddr:     prgAddrHigh,
					InitStackAddr - 1: prgAddrLow},
			},
			State{
				StackPtr:       InitStackPtr,
				ProgramCounter: prgAddr + subroutineOffset,
				Bus: TestBus{
					InitStackAddr:     prgAddrHigh,
					InitStackAddr - 1: prgAddrLow},
			},
		},

		{"SEC_SetCarryFlag", SEC,
			State{Status: status &^ Carry},
			State{Status: status | Carry},
		},

		{"SEI_SetInterruptDisableFlag", SEI,
			State{Status: status &^ InterruptDisable},
			State{Status: status | InterruptDisable},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.cmd(&test.before)
			expectStateEq(t, &test.before, &test.after)
		})
	}
}
