package service

import "fmt"

var ErrAlreadyExists = fmt.Errorf("already exists")
var ErrInternal = fmt.Errorf("internal error")
var ErrNotFound = fmt.Errorf("not found")
var ErrBadCredentials = fmt.Errorf("bad credentials")
var ErrInvalidToken = fmt.Errorf("invalid token")
var ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
var ErrInvalidKID = fmt.Errorf("invalid kid")
