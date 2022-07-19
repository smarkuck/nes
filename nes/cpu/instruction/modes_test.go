package instruction_test

import (
	"github.com/smarkuck/nes/nes/cpu"
	. "github.com/smarkuck/nes/nes/cpu/instruction"
	. "github.com/smarkuck/unittest"
)

const (
	programAddr = 0xc1fe
	value       = 26
	cycles      = 8
	bonus       = 2

	basicShift  = 2
	basicCycles = 2
)

func newState(p program, m memory) *cpu.State {
	return &cpu.State{
		ProgramCounter: programAddr,
		Bus:            newBus(p, m),
	}
}

func newStateX(x byte, p program, m memory) *cpu.State {
	s := newState(p, m)
	s.RegisterX = x
	return s
}

func newStateY(y byte, p program, m memory) *cpu.State {
	s := newState(p, m)
	s.RegisterY = y
	return s
}

type testBus map[uint16]byte
type program = []byte
type memory = map[uint16]byte

func newBus(p program, m memory) testBus {
	bus := testBus{}
	bus.loadProgram(p)
	bus.loadMemory(m)
	return bus
}

func (t testBus) loadProgram(p program) {
	offset := programAddr + 1
	for i, v := range p {
		addr := uint16(offset + i)
		t[addr] = v
	}
}

func (t testBus) loadMemory(m memory) {
	for addr, v := range m {
		t[addr] = v
	}
}

func (t testBus) Read(addr uint16) byte {
	return t[addr]
}

func (t testBus) Write(addr uint16, value byte) {
	t[addr] = value
}

var idleCmd = func(*cpu.State, uint16) {}

type impliedCmd = func(*cpu.State)
type addressCmd = func(_ *cpu.State, addr uint16)
type relativeCmd = func(status byte) bool

func transformToAddressCmd(c impliedCmd) addressCmd {
	return func(s *cpu.State, _ uint16) { c(s) }
}

func Test_OnExecute_RunProvidedCommandOnce(t *T) {
	var counter uint
	count := func(*cpu.State) { counter++ }
	countAddr := transformToAddressCmd(count)
	countRel := func(byte) bool { counter++; return true }

	tests := []struct {
		name        string
		instruction cpu.Instruction
	}{
		{"Implied", NewImplied(count, cycles)},
		{"Accumulative", NewAccumulative(count, cycles)},
		{"Immediate", NewImmediate(countAddr, cycles)},
		{"ZeroPage", NewZeroPage(countAddr, cycles)},
		{"ZeroPageX", NewZeroPageX(countAddr, cycles)},
		{"ZeroPageY", NewZeroPageY(countAddr, cycles)},
		{"Absolute", NewAbsolute(countAddr, cycles)},
		{"AbsoluteX", NewAbsoluteX(countAddr, cycles, bonus)},
		{"AbsoluteY", NewAbsoluteY(countAddr, cycles, bonus)},
		{"Indirect", NewIndirect(countAddr, cycles)},
		{"IndirectX", NewIndirectX(countAddr, cycles)},
		{"IndirectY", NewIndirectY(countAddr, cycles, bonus)},
		{"Relative", NewRelative(countRel)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			counter = 0
			test.instruction.Execute(newState(nil, nil))
			ExpectEq(t, counter, 1)
		})
	}
}

func Test_OnExecute_ShiftProgramCounterBeforeCommand(t *T) {
	var counter uint16
	save := func(s *cpu.State) { counter = s.ProgramCounter }
	saveAddr := transformToAddressCmd(save)

	tests := []struct {
		name        string
		shift       uint16
		instruction cpu.Instruction
	}{
		{"Implied", 1, NewImplied(save, cycles)},
		{"Accumulative", 1, NewAccumulative(save, cycles)},
		{"Immediate", 2, NewImmediate(saveAddr, cycles)},
		{"ZeroPage", 2, NewZeroPage(saveAddr, cycles)},
		{"ZeroPageX", 2, NewZeroPageX(saveAddr, cycles)},
		{"ZeroPageY", 2, NewZeroPageY(saveAddr, cycles)},
		{"Absolute", 3, NewAbsolute(saveAddr, cycles)},
		{"AbsoluteX", 3, NewAbsoluteX(saveAddr, cycles, bonus)},
		{"AbsoluteY", 3, NewAbsoluteY(saveAddr, cycles, bonus)},
		{"Indirect", 3, NewIndirect(saveAddr, cycles)},
		{"IndirectX", 2, NewIndirectX(saveAddr, cycles)},
		{"IndirectY", 2, NewIndirectY(saveAddr, cycles, bonus)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.instruction.Execute(newState(nil, nil))
			ExpectEqf(t,
				counter, programAddr+test.shift, TwoHexBytes)
		})
	}
}

func Test_OnExecute_ProvidedCommandCanModifyState(t *T) {
	save := func(s *cpu.State) { s.Accumulator = value }
	saveAddr := transformToAddressCmd(save)

	tests := []struct {
		name        string
		instruction cpu.Instruction
	}{
		{"Implied", NewImplied(save, cycles)},
		{"Accumulative", NewAccumulative(save, cycles)},
		{"Immediate", NewImmediate(saveAddr, cycles)},
		{"ZeroPage", NewZeroPage(saveAddr, cycles)},
		{"ZeroPageX", NewZeroPageX(saveAddr, cycles)},
		{"ZeroPageY", NewZeroPageY(saveAddr, cycles)},
		{"Absolute", NewAbsolute(saveAddr, cycles)},
		{"AbsoluteX", NewAbsoluteX(saveAddr, cycles, bonus)},
		{"AbsoluteY", NewAbsoluteY(saveAddr, cycles, bonus)},
		{"Indirect", NewIndirect(saveAddr, cycles)},
		{"IndirectX", NewIndirectX(saveAddr, cycles)},
		{"IndirectY", NewIndirectY(saveAddr, cycles, bonus)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			s := newState(nil, nil)
			test.instruction.Execute(s)
			ExpectEq(t, s.Accumulator, value)
		})
	}
}

func Test_OnExecute_PassProperAddressToCommand(t *T) {
	var address uint16
	save := func(_ *cpu.State, addr uint16) { address = addr }

	tests := []struct {
		name         string
		instruction  cpu.Instruction
		expectedAddr uint16
		state        *cpu.State
	}{
		{"Immediate_NextByteAddress",
			NewImmediate(save, cycles),
			programAddr + 1, newState(nil, nil)},

		{"ZeroPage_NextByte",
			NewZeroPage(save, cycles),
			0x00c7, newState(
				program{0xc7}, nil)},

		{"ZeroPageX_ZeroPage+RegisterX",
			NewZeroPageX(save, cycles),
			0x00c9, newStateX(2,
				program{0xc7}, nil)},

		{"ZeroPageX_ByteOverflow",
			NewZeroPageX(save, cycles),
			0x0000, newStateX(1,
				program{0xff}, nil)},

		{"ZeroPageY_ZeroPage+RegisterY",
			NewZeroPageY(save, cycles),
			0x00c9, newStateY(2,
				program{0xc7}, nil)},

		{"ZeroPageY_ByteOverflow",
			NewZeroPageY(save, cycles),
			0x0000, newStateY(1,
				program{0xff}, nil)},

		{"Absolute_NextTwoBytes",
			NewAbsolute(save, cycles),
			0x45c7, newState(
				program{0xc7, 0x45}, nil)},

		{"AbsoluteX_Absolute+RegisterX",
			NewAbsoluteX(save, cycles, bonus),
			0x46c6, newStateX(255,
				program{0xc7, 0x45}, nil)},

		{"AbsoluteY_Absolute+RegisterY",
			NewAbsoluteY(save, cycles, bonus),
			0x46c6, newStateY(255,
				program{0xc7, 0x45}, nil)},

		{"Indirect_TwoBytesFromAbsolute",
			NewIndirect(save, cycles),
			0x46c6, newState(
				program{0xc7, 0x45},
				memory{0x45c7: 0xc6, 0x45c8: 0x46})},

		{"Indirect_CPUPageOverflowBug",
			NewIndirect(save, cycles),
			0x46c6, newState(
				program{0xff, 0x45},
				memory{0x45ff: 0xc6, 0x4500: 0x46})},

		{"IndirectX_TwoBytesFromZeroPageX",
			NewIndirectX(save, cycles),
			0x46c6, newStateX(2,
				program{0xc7},
				memory{0x00c9: 0xc6, 0x00ca: 0x46})},

		{"IndirectX_ByteOverflow",
			NewIndirectX(save, cycles),
			0x46c6, newStateX(1,
				program{0xff},
				memory{0x0000: 0xc6, 0x0001: 0x46})},

		{"IndirectX_PageOverflow",
			NewIndirectX(save, cycles),
			0x46c6, newStateX(1,
				program{0xfe},
				memory{0x00ff: 0xc6, 0x0000: 0x46})},

		{"IndirectY_TwoBytesFromZeroPage+RegisterY",
			NewIndirectY(save, cycles, bonus),
			0x47c5, newStateY(255,
				program{0xc7},
				memory{0x00c7: 0xc6, 0x00c8: 0x46})},

		{"IndirectY_PageOverflow",
			NewIndirectY(save, cycles, bonus),
			0x47c5, newStateY(255,
				program{0xff},
				memory{0x00ff: 0xc6, 0x0000: 0x46})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.instruction.Execute(test.state)
			ExpectEqf(t,
				address, test.expectedAddr, TwoHexBytes)
		})
	}
}

func Test_RelativeMode_OnExecute_ShiftProgramCounter(t *T) {
	cmd := func(status byte) bool { return status > 0 }

	tests := []struct {
		name         string
		status       byte
		shift        byte
		expectedAddr uint16
	}{
		{"FalsePredicate", 0, 0x7f, programAddr + basicShift},
		{"Basic", 1, 0x00, programAddr + basicShift},
		{"PlusOne", 1, 0x01, programAddr + basicShift + 1},
		{"PlusMax", 1, 0x7f, programAddr + basicShift + 127},
		{"MinusOne", 1, 0xff, programAddr + basicShift - 1},
		{"MinusMax", 1, 0x80, programAddr + basicShift - 128},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			s := newState(program{test.shift}, nil)
			s.Status = test.status

			NewRelative(cmd).Execute(s)

			ExpectEqf(t,
				s.ProgramCounter, test.expectedAddr,
				TwoHexBytes)
		})
	}
}

func Test_OnGetCycles_ReturnBasicProvidedNumOfCycles(t *T) {
	tests := []struct {
		name        string
		instruction cpu.Instruction
	}{
		{"Implied", NewImplied(nil, cycles)},
		{"Accumulative", NewAccumulative(nil, cycles)},
		{"Immediate", NewImmediate(nil, cycles)},
		{"ZeroPage", NewZeroPage(nil, cycles)},
		{"ZeroPageX", NewZeroPageX(nil, cycles)},
		{"ZeroPageY", NewZeroPageY(nil, cycles)},
		{"Absolute", NewAbsolute(nil, cycles)},
		{"AbsoluteX", NewAbsoluteX(nil, cycles, bonus)},
		{"AbsoluteY", NewAbsoluteY(nil, cycles, bonus)},
		{"Indirect", NewIndirect(nil, cycles)},
		{"IndirectX", NewIndirectX(nil, cycles)},
		{"IndirectY", NewIndirectY(nil, cycles, bonus)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			ExpectEq(t, test.instruction.GetCycles(), cycles)
		})
	}
}

func Test_OnExecute_DontIncreaseCyclesIfPageNotCrossed(t *T) {
	tests := []struct {
		name        string
		instruction cpu.Instruction
		state       *cpu.State
	}{
		{"AbsoluteX", NewAbsoluteX(idleCmd, cycles, bonus),
			newStateX(1, program{0xfe, 0x45}, nil)},

		{"AbsoluteY", NewAbsoluteY(idleCmd, cycles, bonus),
			newStateY(1, program{0xfe, 0x45}, nil)},

		{"IndirectY", NewIndirectY(idleCmd, cycles, bonus),
			newStateY(1, program{0xa2},
				memory{0x00a2: 0xfe, 0x00a3: 0x45})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.instruction.Execute(test.state)
			ExpectEq(t, test.instruction.GetCycles(), cycles)
		})
	}
}

func Test_OnExecute_IncreaseCyclesOnlyOnceIfPageCrossed(t *T) {
	tests := []struct {
		name  string
		instr cpu.Instruction
		state *cpu.State
	}{
		{"AbsoluteX", NewAbsoluteX(idleCmd, cycles, bonus),
			newStateX(1, program{0xff, 0x45}, nil)},

		{"AbsoluteY", NewAbsoluteY(idleCmd, cycles, bonus),
			newStateY(1, program{0xff, 0x45}, nil)},

		{"IndirectY", NewIndirectY(idleCmd, cycles, bonus),
			newStateY(1, program{0xc7},
				memory{0x00c7: 0xff, 0x00c8: 0x45})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.instr.Execute(test.state)
			ExpectEq(t, test.instr.GetCycles(), cycles+bonus)

			test.state.ProgramCounter = programAddr
			test.instr.Execute(test.state)
			ExpectEq(t, test.instr.GetCycles(), cycles+bonus)
		})
	}
}

func Test_OnExecute_DecreaseCyclesIfPageNotCrossed(t *T) {
	tests := []struct {
		name  string
		instr cpu.Instruction
		state *cpu.State
	}{
		{"AbsoluteX", NewAbsoluteX(idleCmd, cycles, bonus),
			newStateX(1, program{0xff, 0x45}, nil)},

		{"AbsoluteY", NewAbsoluteY(idleCmd, cycles, bonus),
			newStateY(1, program{0xff, 0x45}, nil)},

		{"IndirectY", NewIndirectY(idleCmd, cycles, bonus),
			newStateY(1, program{0xc7},
				memory{0x00c7: 0xff, 0x00c8: 0x45})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			test.instr.Execute(test.state)
			ExpectEq(t, test.instr.GetCycles(), cycles+bonus)

			test.state.ProgramCounter = programAddr
			test.state.RegisterX, test.state.RegisterY = 0, 0
			test.instr.Execute(test.state)
			ExpectEq(t, test.instr.GetCycles(), cycles)
		})
	}
}

func Test_RelativeMode_ReturnTwoCyclesByDefault(t *T) {
	ExpectEq(t, NewRelative(nil).GetCycles(), basicCycles)
}

func Test_RelativeMode_ReturnTwoCyclesIfCmdReturnFalse(t *T) {
	cmd := func(byte) bool { return false }
	i := NewRelative(cmd)

	i.Execute(newState(nil, nil))

	ExpectEq(t, i.GetCycles(), basicCycles)
}

func Test_RelativeMode_RtnCyclesBasedOnCmdAndPageCross(t *T) {
	var cmdResult bool
	cmd := func(byte) bool { return cmdResult }

	type phase struct {
		cmdResult   bool
		shift       byte
		bonusCycles uint8
	}

	tests := []struct {
		name   string
		phases [2]phase
	}{
		{"ShiftWithoutCross_NoShift",
			[2]phase{{true, 0x00, 1}, {false, 0x80, 0}}},
		{"ShiftWithCross_NoShift",
			[2]phase{{true, 0xff, 2}, {false, 0x80, 0}}},
		{"TwoShiftsWithoutCross",
			[2]phase{{true, 0x00, 1}, {true, 0x00, 1}}},
		{"TwoShiftsWithCross",
			[2]phase{{true, 0xff, 2}, {true, 0xff, 2}}},
		{"ShiftWithoutCross_ShiftWithCross",
			[2]phase{{true, 0x00, 1}, {true, 0xff, 2}}},
		{"ShiftWithCross_ShiftWithoutCross",
			[2]phase{{true, 0xff, 2}, {true, 0x00, 1}}},
	}

	testPhase := func(t *T, i cpu.Instruction, p phase) {
		cmdResult = p.cmdResult
		s := newState(program{p.shift}, nil)
		i.Execute(s)
		ExpectEq(t, i.GetCycles(), basicCycles+p.bonusCycles)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			i := NewRelative(cmd)
			testPhase(t, i, test.phases[0])
			testPhase(t, i, test.phases[1])
		})
	}
}
