package evaluator

import (
	"context"
	"fmt"
)

type Decision int

const (
	MaintainCapacity Decision = iota
	IncreaseCapacity
	ReduceCapacity
)

func (d Decision) String() string {
	switch d {
	case MaintainCapacity:
		return "MaintainCapacity"
	case IncreaseCapacity:
		return "IncreaseCapacity"
	case ReduceCapacity:
		return "ReduceCapacity"
	default:
		return fmt.Sprintf("Decision(%d)", int(d))
	}
}

type Rule interface {
	Evaluate(ctx context.Context, currentCapacity int, resourceIDs []string) (Decision, int, string, error)
}
