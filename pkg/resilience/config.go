package resilience

import (
	"errors"

	"github.com/maisiq/go-auth-service/internal/repository"
)

type ClientType uint8

const (
	SqlDBClient ClientType = iota
	InMemoryDBClient
	HTTPClient
)

func IsDBFailure(err error) bool {
	return err == nil ||
		errors.Is(err, repository.ErrNotFound) ||
		errors.Is(err, repository.ErrAlreadyExists)
}

func IsInMemoryDBFailure(err error) bool {
	return err == nil ||
		errors.Is(err, repository.ErrNotFound)
}

func IsHTTPClientFailure(err error) bool {
	return err == nil
}

// filter out domain errors
func getIsBusinessErrFilter(clientType ClientType) func(error) bool {
	var fn func(error) bool

	switch clientType {
	case SqlDBClient:
		fn = IsDBFailure
	case InMemoryDBClient:
		fn = IsInMemoryDBFailure
	case HTTPClient:
		fn = IsHTTPClientFailure
	default:
		panic("resilience.getErrFilter: unregistred client type")
	}
	return fn
}
