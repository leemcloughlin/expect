// +build windows

package expect

import (
	"context"
	"errors"
)

const (
	// No good way to discover EOF so have to take a best guess
	// Default (above) to control-D
	EOF = "\032" // control-Z
)

func newExpectCommon(parentCtx context.Context, reap bool, prog string, arg ...string) (exp *Expect, err error) {
	return nil, errors.New("expect is not yet implemented on Windows - sorry")
}
