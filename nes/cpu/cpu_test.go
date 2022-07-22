package cpu_test

import (
	"fmt"

	. "github.com/smarkuck/nes/nes/cpu"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	"github.com/smarkuck/nes/nes/cpu/state"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	resetPrgAddr = 0x1050
	address      = 0x1060
	code         = 0x07
	value        = 0xea
	cycles       = 13

	invalidRemainingCyclesText = "invalid remaining cycles"
	invalidExecCountText       = "invalid number of executions"
	invalidErrorText           = "invalid error message"
	invalidBusValueText        = "invalid value in bus"

	unknownInstrFormat = "unknown instruction code: " +
		byteutil.HexByte

	invalidCyclesFormat = "encountered instruction needs " +
		"0 cycles to execute: " + byteutil.HexByte
)

type execChecker struct {
	cycles    uint8
	execCount uint
}

func (e *execChecker) Execute(*state.State) {
	e.execCount++
}

func (e *execChecker) GetCycles() uint8 {
	return e.cycles
}

func (e execChecker) expectExecCountEq(t *T, value uint) {
	ExpectEq(t, e.execCount, value, invalidExecCountText)
}

type addressIncrementer struct {
	address uint16
}

func (a addressIncrementer) Execute(s *state.State) {
	v := s.Read(a.address)
	s.Write(a.address, v+1)
}

func (addressIncrementer) GetCycles() uint8 {
	return cycles
}

type stateModifier struct {
	value byte
}

func (s stateModifier) Execute(state *state.State) {
	*state = *NewState(s.value, state.Bus)
}

func (stateModifier) GetCycles() uint8 {
	return cycles
}

func expectRemainingCyclesEq(t *T, cpu CPU, value uint8) {
	ExpectEq(t, cpu.GetRemainingCycles(), value,
		invalidRemainingCyclesText)
}

func getUnknownInstrText(code byte) string {
	return fmt.Sprintf(unknownInstrFormat, code)
}

func getInvalidCyclesText(code byte) string {
	return fmt.Sprintf(invalidCyclesFormat, code)
}

func Test_OnNewCPU_CreateCPUWithProperState(t *T) {
	bus := TestBus{}
	cpu := NewCPU(bus, nil)

	ExpectStateEq(t, cpu.GetState(), NewInitState(0, bus))
	expectRemainingCyclesEq(t, cpu, 0)
}

func Test_CannotManipulateCPUByReturnedState(t *T) {
	cpu := NewCPU(TestBus{}, nil)

	cpu.GetState().Accumulator = 1

	ExpectAccumulatorEq(t, cpu.GetState(), 0)
}

type cpuSuite struct {
	bus TestBus
}

func (s *cpuSuite) Setup() {
	s.bus = NewTestBusResetPrg(resetPrgAddr, code)
}

func (s cpuSuite) newCPU(i Instructions) CPU {
	return NewCPU(s.bus, i)
}

func (s cpuSuite) expectBusValueEq(t *T,
	addr uint16, value byte) {
	ExpectEqf(t, s.bus[addr], value,
		byteutil.HexByte, invalidBusValueText)
}

func Test_CPU(t *T) {
	TestSuite(t, new(cpuSuite))
}

func (s cpuSuite) OnNewCPU_LoadProgramFromResetVector(t *T) {
	cpu := s.newCPU(nil)

	ExpectProgramCounterEq(t, cpu.GetState(), resetPrgAddr)
}

func (s cpuSuite) WhenNoCyclesLeft_ExecNextInstr(t *T) {
	checker := &execChecker{cycles: cycles}
	cpu := s.newCPU(Instructions{code: checker})

	cpu.Tick()

	checker.expectExecCountEq(t, 1)
	expectRemainingCyclesEq(t, cpu, cycles-1)
}

func (s cpuSuite) WhenInstructionIsUnknown_Panic(t *T) {
	cpu := s.newCPU(Instructions{})

	defer ExpectPanicErrEq(t,
		getUnknownInstrText(code), invalidErrorText)

	cpu.Tick()
}

func (s cpuSuite) WhenInstrReturnsZeroCycles_Panic(t *T) {
	c := &execChecker{cycles: 0}
	cpu := s.newCPU(Instructions{code: c})

	defer ExpectPanicErrEq(t,
		getInvalidCyclesText(code), invalidErrorText)

	cpu.Tick()
}

func (s cpuSuite) WhenMoreThanZeroCycles_SkipCycle(t *T) {
	checker := &execChecker{cycles: cycles}
	cpu := s.newCPU(Instructions{code: checker})

	cpu.Tick()
	cpu.Tick()

	checker.expectExecCountEq(t, 1)
	expectRemainingCyclesEq(t, cpu, cycles-2)
}

func (s cpuSuite) InstructionCanInteractWithBus(t *T) {
	a := addressIncrementer{address}
	cpu := s.newCPU(Instructions{code: a})
	s.bus[address] = value

	cpu.Tick()

	s.expectBusValueEq(t, address, value+1)
}

func (s cpuSuite) InstructionCanModifyCPUState(t *T) {
	sm := stateModifier{value}
	cpu := s.newCPU(Instructions{code: sm})

	cpu.Tick()

	ExpectStateEq(t,
		cpu.GetState(), NewState(value, s.bus))
}

func (s cpuSuite) OnReset_RestoreInitialState(t *T) {
	sm := stateModifier{value}
	cpu := s.newCPU(Instructions{code: sm})

	cpu.Tick()
	cpu.Reset()

	ExpectStateEq(t,
		cpu.GetState(), NewInitState(resetPrgAddr, s.bus))
	expectRemainingCyclesEq(t, cpu, 0)
}
