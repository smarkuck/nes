package cpu_test

import (
	"github.com/smarkuck/nes/nes/cpu"
	. "github.com/smarkuck/nes/nes/cpu/testutil"
	. "github.com/smarkuck/unittest"
)

const (
	addToA, addToACycles                     = 0x65, 3
	branchIfYNotZero, branchIfYNotZeroCycles = 0xd0, 2
	clearCarry, clearCarryCycles             = 0x18, 2
	decrementY, decrementYCycles             = 0x88, 2
	loadA, loadACycles                       = 0xa9, 2
	loadX, loadXCycles                       = 0xa2, 2
	loadY, loadYCycles                       = 0xa4, 3
	storeA, storeACycles                     = 0x85, 3
	storeX, storeXCycles                     = 0x86, 3
	bonusBranchCycle                         = 1

	prgCycles = loadXCycles +
		storeXCycles +
		loadXCycles +
		storeXCycles +
		loadYCycles +
		loadACycles +
		clearCarryCycles +
		value2*(addToACycles+
			decrementYCycles+
			branchIfYNotZeroCycles+
			bonusBranchCycle) - bonusBranchCycle +
		storeACycles

	prgAddr        = 0x8000
	resultAddr     = 0x02
	jumpToAddition = 0xfb
	value1Addr     = 0x00
	value2Addr     = 0x01
	value1         = 3
	value2         = 10
)

func getProgram() Program {
	return Program{
		loadX, value1,
		storeX, value1Addr,
		loadX, value2,
		storeX, value2Addr,
		loadY, value2Addr,
		loadA, 0,
		clearCarry,
		addToA, value1Addr,
		decrementY,
		branchIfYNotZero, jumpToAddition,
		storeA, resultAddr,
	}
}

func Test_CPU6502_MultiplicationProgram(t *T) {
	prg := getProgram()
	prgLen := uint16(len(prg))
	bus := NewTestBusResetPrg(prgAddr, prg)
	cpu := cpu.NewCPU6502(bus)

	for i := 0; i < prgCycles; i++ {
		cpu.Tick()
	}

	ExpectEq(t, bus[value1Addr], value1)
	ExpectEq(t, bus[value2Addr], value2)
	ExpectEq(t, bus[resultAddr], value1*value2)
	ExpectProgramCounterEq(t, cpu.GetState(), prgAddr+prgLen)
}

func Test_CPU6502_StartNextInstructionAfterProgram(t *T) {
	bus := NewTestBusResetPrg(prgAddr, getProgram())
	cpu := cpu.NewCPU6502(bus)

	for i := 0; i < prgCycles+1; i++ {
		cpu.Tick()
	}

	ExpectProgramCounterEq(t, cpu.GetState(), 0x0000)
}
