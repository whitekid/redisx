package redisx

import "github.com/whitekid/goxp/errors"

var ErrNotFound = errors.New("not found")
var ErrNoValue = errors.New("no value")

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
func IsNoValue(err error) bool  { return errors.Is(err, ErrNoValue) }
