package resilience

import (
	"math/rand/v2"
	"time"

	"github.com/sony/gobreaker"
)

func NewCircuitBreakerDecorator(name string, clientType ClientType) *gobreaker.CircuitBreaker {
	filter := getIsBusinessErrFilter(clientType)
	timeout := 30 + time.Duration(rand.Int64N(int64(10)))

	return gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name: name,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
			IsSuccessful: filter,
			Timeout:      timeout,
		},
	)
}
