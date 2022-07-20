package instruction

import "github.com/smarkuck/nes/nes/cpu/state"

type Instruction interface {
	Execute(*State)
	GetCycles() uint8
}

type State = state.State
