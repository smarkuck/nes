package cpu_test

import (
	"fmt"

	. "github.com/smarkuck/nes/nes/cpu"
	"github.com/smarkuck/nes/nes/cpu/byteutil"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	initStatus       = 0x34
	initStackPtr     = 0xfd
	resetVector      = 0xfffc
	resetProgramAddr = 0x1050

	address = 0x1060
	code    = 0x07
	cycles  = 13
	value   = 234

	invalidAccumulatorText     = "invalid accumulator"
	invalidRegisterXText       = "invalid register X"
	invalidRegisterYText       = "invalid register Y"
	invalidStatusText          = "invalid status"
	invalidStackPtrText        = "invalid stack pointer"
	invalidProgramCounterText  = "invalid program counter"
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

func (e *execChecker) Execute(*State) {
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

func (a addressIncrementer) Execute(s *State) {
	v := s.Read(a.address)
	s.Write(a.address, v+1)
}

func (addressIncrementer) GetCycles() uint8 {
	return cycles
}

type stateModifier struct {
	value byte
}

func (s stateModifier) Execute(state *State) {
	state.Accumulator = s.value
	state.RegisterX = s.value
	state.RegisterY = s.value
	state.Status = s.value
	state.StackPtr = s.value
	state.ProgramCounter = uint16(s.value)
}

func (stateModifier) GetCycles() uint8 {
	return cycles
}

func expectAccumulatorEq(t *T, cpu CPU, value byte) {
	ExpectEq(t, cpu.GetAccumulator(), value,
		invalidAccumulatorText)
}

func expectRegisterXEq(t *T, cpu CPU, value byte) {
	ExpectEq(t, cpu.GetRegisterX(), value,
		invalidRegisterXText)
}

func expectRegisterYEq(t *T, cpu CPU, value byte) {
	ExpectEq(t, cpu.GetRegisterY(), value,
		invalidRegisterYText)
}

func expectRegistersEq(t *T, cpu CPU, value byte) {
	expectAccumulatorEq(t, cpu, value)
	expectRegisterXEq(t, cpu, value)
	expectRegisterYEq(t, cpu, value)
}

func expectStatusEq(t *T, cpu CPU, value byte) {
	ExpectEqf(t, cpu.GetStatus(), value,
		byteutil.BinByte, invalidStatusText)
}

func expectStackPtrEq(t *T, cpu CPU, value byte) {
	ExpectEqf(t, cpu.GetStackPtr(), value,
		byteutil.HexByte, invalidStackPtrText)
}

func expectProgramCounterEq(t *T, cpu CPU, value uint16) {
	ExpectEqf(t, cpu.GetProgramCounter(), value,
		byteutil.TwoHexBytes, invalidProgramCounterText)
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
	cpu := NewCPU(TestBus{}, Instructions{})

	expectRegistersEq(t, cpu, 0)
	expectStatusEq(t, cpu, initStatus)
	expectStackPtrEq(t, cpu, initStackPtr)
	expectRemainingCyclesEq(t, cpu, 0)
}

type cpuSuite struct {
	bus TestBus
}

func (s *cpuSuite) Setup() {
	s.bus = NewTestBus(resetVector,
		Program{
			byteutil.GetLow(resetProgramAddr),
			byteutil.GetHigh(resetProgramAddr),
		},
		Memory{resetProgramAddr: code})
}

func (s cpuSuite) newCPU(i Instructions) CPU {
	return NewCPU(s.bus, i)
}

func (s cpuSuite) expectBusValueEq(t *T,
	addr uint16, value byte) {
	ExpectEq(t, s.bus[addr], value, invalidBusValueText)
}

func Test_CPU(t *T) {
	TestSuite(t, new(cpuSuite))
}

func (s cpuSuite) OnNewCPU_LoadProgramFromResetVector(t *T) {
	cpu := s.newCPU(Instructions{})

	expectProgramCounterEq(t, cpu, resetProgramAddr)
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

	expectRegistersEq(t, cpu, value)
	expectStatusEq(t, cpu, value)
	expectStackPtrEq(t, cpu, value)
	expectProgramCounterEq(t, cpu, value)
}

func (s cpuSuite) OnReset_RestoreInitialState(t *T) {
	sm := stateModifier{value}
	cpu := s.newCPU(Instructions{code: sm})

	cpu.Tick()
	cpu.Reset()

	expectRegistersEq(t, cpu, 0)
	expectStatusEq(t, cpu, initStatus)
	expectStackPtrEq(t, cpu, initStackPtr)
	expectProgramCounterEq(t, cpu, resetProgramAddr)
	expectRemainingCyclesEq(t, cpu, 0)
}
