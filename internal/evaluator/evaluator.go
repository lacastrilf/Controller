package evaluator

import "context"

type Decision int

const (
	MaintainCapacity Decision = iota
	IncreaseCapacity
	ReduceCapacity
)

type Rule interface {
	Evaluate(ctx context.Context, currentCapacity int, resourceIDs []string) (Decision, int, string, error)
}
