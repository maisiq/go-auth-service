package resilience

import (
	"time"

	"github.com/avast/retry-go/v4"
)

type RetryConfig struct {
	Client      ClientType
	MaxAttempts uint
	MaxDelay    time.Duration
	filter      func(error) bool
}
type Retry struct {
	cfg RetryConfig
}

func NewRetryDecorator(cfg RetryConfig) *Retry {
	filter := func(err error) bool {
		filter := getIsBusinessErrFilter(cfg.Client)
		return !filter(err)
	}

	cfg.filter = filter
	return &Retry{
		cfg: cfg,
	}
}

func (d *Retry) Call(fn func() error) error {
	return retry.Do(
		fn,
		retry.RetryIf(d.cfg.filter),
		retry.Attempts(d.cfg.MaxAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.MaxDelay(d.cfg.MaxDelay),
		retry.LastErrorOnly(true),
	)
}
