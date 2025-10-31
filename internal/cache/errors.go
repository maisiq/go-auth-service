package cache

import "fmt"

var ErrNotFound error = fmt.Errorf("not found")
var ErrInvalidData error = fmt.Errorf("invalid data")
