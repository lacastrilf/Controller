package controller

import (
	"context"
	"fmt"
	"time"

	"controller/internal/evaluator"
	"controller/internal/executor"
	"controller/internal/logger"
)

type Runner struct {
	ModeName string
	Rule     evaluator.Rule
	Executor executor.Executor
	Logger   *logger.Logger
	Interval time.Duration

	// Cooldown is the minimum time between scaling actions. While it hasn't
	// elapsed since the last successful ScaleUp/ScaleDown, the runner keeps
	// evaluating and logging but does not act — this stops it from reacting
	// to a freshly-launched instance that doesn't have real metrics yet.
	Cooldown time.Duration

	MinCapacity int
	MaxCapacity int

	lastScaleAt time.Time // zero value means "no cooldown in effect yet"
}

func (runner *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(runner.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runner.tick(ctx)
		}
	}
}

func (runner *Runner) tick(ctx context.Context) {
	currentCapacity, err := runner.Executor.CurrentCapacity(ctx)
	if err != nil {
		return
	}

	resourceIDs, err := runner.Executor.CurrentResourceIDs(ctx)
	if err != nil {
		return
	}

	decision, amount, justification, err := runner.Rule.Evaluate(ctx, currentCapacity, resourceIDs)
	if err != nil {
		runner.Logger.Log(logger.Record{
			Timestamp:        time.Now(),
			Mode:             runner.ModeName,
			ExistingCapacity: currentCapacity,
			Decision:         evaluator.MaintainCapacity,
			Justification:    "evaluation error: " + err.Error(),
			ActionResult:     "skipped",
		})
		return
	}

	var actionResult string
	if remaining := runner.cooldownRemaining(); remaining > 0 {
		actionResult = fmt.Sprintf("skipped: cooling down (%s remaining)", remaining.Round(time.Second))
	} else {
		actionResult = runner.act(ctx, decision, amount, currentCapacity)
		if actionResult == "ok" {
			runner.lastScaleAt = time.Now()
		}
	}

	runner.Logger.Log(logger.Record{
		Timestamp:        time.Now(),
		Mode:             runner.ModeName,
		ExistingCapacity: currentCapacity,
		Decision:         decision,
		Justification:    justification,
		ActionResult:     actionResult,
	})
}

// cooldownRemaining returns how much longer the runner must wait before it
// is allowed to scale again, or 0 if it's free to act.
func (runner *Runner) cooldownRemaining() time.Duration {
	if runner.lastScaleAt.IsZero() || runner.Cooldown <= 0 {
		return 0
	}
	remaining := runner.Cooldown - time.Since(runner.lastScaleAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (runner *Runner) act(ctx context.Context, decision evaluator.Decision, amount, currentCapacity int) string {
	switch decision {
	case evaluator.IncreaseCapacity:
		if currentCapacity+amount > runner.MaxCapacity {
			amount = runner.MaxCapacity - currentCapacity
		}
		if amount <= 0 {
			return "skipped: already at max capacity"
		}
		if err := runner.Executor.ScaleUp(ctx, amount); err != nil {
			return "error: " + err.Error()
		}
		return "ok"

	case evaluator.ReduceCapacity:
		if currentCapacity-amount < runner.MinCapacity {
			amount = currentCapacity - runner.MinCapacity
		}
		if amount <= 0 {
			return "skipped: already at min capacity"
		}
		if err := runner.Executor.ScaleDown(ctx, amount); err != nil {
			return "error: " + err.Error()
		}
		return "ok"

	default:
		return "none"
	}
}
