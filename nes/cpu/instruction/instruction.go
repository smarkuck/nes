package instruction

import "github.com/smarkuck/nes/nes/cpu/state"

type Instruction interface {
	Execute(*state.State)
	GetCycles() uint8
}
