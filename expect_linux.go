// +build linux

package expect

import (
	"bytes"
	"context"
	"os/exec"
	"syscall"

	"github.com/kr/pty"
)

const (
	// EOF is the terminal EOF character - NOT guaranteed to be correct on all
	// systems nor if the program passed to expect connects to a different
	// platform
	EOF = "\004"
)

// newExpectCommon uses @parentCtx as cancellation context and @reap to indicate
// whether the spawned process should automatically be reaped.
func newExpectCommon(parentCtx context.Context, reap bool, prog string, arg ...string) (exp *Expect, err error) {
	defer func() {
		// On an error I want to kill the process - if it was started
		if err != nil && exp != nil && exp.cmd.Process != nil {
			debugf("killing process %q due to error (%s)", prog, err)
			exp.Kill()
		}
	}()

	exp = &Expect{
		Buffer:  new(bytes.Buffer),
		bytesIn: make(chan byteIn, ExpectInSize),
	}

	// Create a cancelable child context for exp.cmd (go >= 1.7)
	exp.ctx, exp.cancel = context.WithCancel(parentCtx)
	exp.cmd = exec.CommandContext(exp.ctx, prog, arg...)

	if exp.file, err = pty.Start(exp.cmd); err != nil {
		return nil, err
	}

	// make the pty non blocking so when I read from it I dont jam up
	if err = syscall.SetNonblock(int(exp.file.Fd()), true); err != nil {
		return nil, err
	}
	exp.SetCmdOut(nil)

	go exp.expectReader()

	if reap {
		go exp.expectReaper()
	}

	return exp, nil
}
