package repository

import "fmt"

var ErrAlreadyExists = fmt.Errorf("already exists")
var ErrNotFound = fmt.Errorf("not found")
