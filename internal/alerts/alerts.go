package alerts

import "context"

// Rule defines a simple threshold-based alert rule.
type Rule struct {
	Name      string
	Threshold float64
}

// AlertEngine is responsible for evaluating rules and emitting alerts.
type AlertEngine interface {
	Evaluate(ctx context.Context, rule Rule, value float64) (bool, error)
	Close() error
}

type noopEngine struct{}

func NewNoopEngine() AlertEngine { return &noopEngine{} }
func (n *noopEngine) Evaluate(ctx context.Context, rule Rule, value float64) (bool, error) {
	return value > rule.Threshold, nil
}
func (n *noopEngine) Close() error { return nil }
