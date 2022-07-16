package nes_test

import (
	"fmt"

	. "github.com/smarkuck/nes/nes"
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
	kb      = 0x400

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
	unknownInstrFormat         = "unknown instruction code: " + HexByte
	invalidCyclesFormat        = "encountered instruction needs " +
		"0 cycles to execute: " + HexByte
)

type testBus [64 * kb]byte

func (t *testBus) Read(addr uint16) byte {
	return t[addr]
}

func (t *testBus) Write(addr uint16, value byte) {
	t[addr] = value
}

type execChecker struct {
	cycles    uint8
	execCount uint
}

func (e *execChecker) Execute(*CPUState) {
	e.execCount++
}

func (e execChecker) GetCycles() uint8 {
	return e.cycles
}

func (e execChecker) expectExecCountEq(t *T, value uint) {
	ExpectEq(t, e.execCount, value, invalidExecCountText)
}

type addressIncrementer struct {
	address uint16
}

func (a addressIncrementer) Execute(s *CPUState) {
	v := s.Read(a.address)
	s.Write(a.address, v+1)
}

func (addressIncrementer) GetCycles() uint8 {
	return cycles
}

type stateModifier struct {
	value byte
}

func (s stateModifier) Execute(state *CPUState) {
	state.Accumulator = s.value
	state.RegisterX = s.value
	state.RegisterY = s.value
	state.Status = s.value
	state.StackPtr = s.value
	state.ProgramCounter = uint16(s.value)
}

func (s stateModifier) GetCycles() uint8 {
	return s.value
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
		BinByte, invalidStatusText)
}

func expectStackPtrEq(t *T, cpu CPU, value byte) {
	ExpectEqf(t, cpu.GetStackPtr(), value,
		HexByte, invalidStackPtrText)
}

func expectProgramCounterEq(t *T, cpu CPU, value uint16) {
	ExpectEqf(t, cpu.GetProgramCounter(), value,
		TwoHexBytes, invalidProgramCounterText)
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
	cpu := NewCPU(new(testBus), Instructions{})

	expectRegistersEq(t, cpu, 0)
	expectStatusEq(t, cpu, initStatus)
	expectStackPtrEq(t, cpu, initStackPtr)
	expectRemainingCyclesEq(t, cpu, 0)
}

type Suite struct {
	bus *testBus
}

func (s *Suite) Setup() {
	s.bus = new(testBus)
	s.bus[resetVector] = getLowByte(resetProgramAddr)
	s.bus[resetVector+1] = getHighByte(resetProgramAddr)
	s.bus[resetProgramAddr] = code
}

func getLowByte(value uint16) byte {
	return byte(value)
}

func getHighByte(value uint16) byte {
	return byte(value >> 8)
}

func (s *Suite) newCPU(i Instructions) CPU {
	return NewCPU(s.bus, i)
}

func (s *Suite) expectBusValueEq(t *T, addr uint16, value byte) {
	ExpectEq(t, s.bus[addr], value, invalidBusValueText)
}

func Test_CPU(t *T) {
	TestSuite(t, new(Suite))
}

func (s *Suite) OnNewCPU_LoadProgramFromResetVector(t *T) {
	cpu := s.newCPU(Instructions{})

	expectProgramCounterEq(t, cpu, resetProgramAddr)
}

func (s *Suite) WhenNoCyclesLeft_ExecNextInstruction(t *T) {
	checker := &execChecker{cycles: cycles}
	cpu := s.newCPU(Instructions{code: checker})

	cpu.Tick()

	checker.expectExecCountEq(t, 1)
	expectRemainingCyclesEq(t, cpu, cycles-1)
}

func (s *Suite) WhenInstructionIsUnknown_Panic(t *T) {
	cpu := s.newCPU(Instructions{})

	defer ExpectPanicErrEq(t,
		getUnknownInstrText(code), invalidErrorText)

	cpu.Tick()
}

func (s *Suite) WhenInstructionReturnsZeroCycles_Panic(t *T) {
	c := &execChecker{cycles: 0}
	cpu := s.newCPU(Instructions{code: c})

	defer ExpectPanicErrEq(t,
		getInvalidCyclesText(code), invalidErrorText)

	cpu.Tick()
}

func (s *Suite) WhenMoreThanZeroCycles_SkipCycle(t *T) {
	checker := &execChecker{cycles: cycles}
	cpu := s.newCPU(Instructions{code: checker})

	cpu.Tick()
	cpu.Tick()

	checker.expectExecCountEq(t, 1)
	expectRemainingCyclesEq(t, cpu, cycles-2)
}

func (s *Suite) InstructionCanInteractWithBus(t *T) {
	a := addressIncrementer{address}
	cpu := s.newCPU(Instructions{code: a})
	s.bus[address] = value

	cpu.Tick()

	s.expectBusValueEq(t, address, value+1)
}

func (s *Suite) InstructionCanModifyCPUState(t *T) {
	sm := stateModifier{value}
	cpu := s.newCPU(Instructions{code: sm})

	cpu.Tick()

	expectRegistersEq(t, cpu, value)
	expectStatusEq(t, cpu, value)
	expectStackPtrEq(t, cpu, value)
	expectProgramCounterEq(t, cpu, value)
}

func (s *Suite) OnReset_RestoreInitialState(t *T) {
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
