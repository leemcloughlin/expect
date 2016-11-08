/*
File summary: Simple version of expect in pure Go
Package: expect
Author: Lee McLoughlin

Copyright (C) 2016 LMMR Tech Ltd

*/

/*
Expect is pure Go (golang) version of the terminal interaction package Expect
common on many Linux systems

A very simple example that calls the rev program to reverse the text in each
line is:

	exp, err := NewExpect("rev")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(5)
	exp.Send("hello\r")

	// i will be the index of the argument that matched the input otherwise
	// i will be a negative number showing match fail or error
	i, found, err := exp.Expect("olleh")
	if i == 0 {
		fmt.Println("found", string(found))
	} else {
		fmt.Println("failed ", err)
	}
	exp.Send(EOF)

This package has only been tested on Linux
*/
package expect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/kr/pty"
)

const (
	// NotFound is returned by Expect for non-matching non-errors (include EOF)
	NotFound = -1

	// TimedOut is returned by Expect on a timeout (along with an error of ETimedOut)
	TimedOut = -2

	// NotStringOrRexgexp is returned if a paramter is not a string or a regexp
	NotStringOrRexgexp = -3
)

var (
	ETimedOut           = errors.New("TimedOut")
	ENotStringOrRexgexp = errors.New("Not string or regexp")
	EReadError          = errors.New("Read Error")

	// Debug if true will generate vast amounts of internal logging
	Debug = false

	// ExpectInSize is the size of the channel between the expectReader and Expect.
	// If you overflow this than expectReader will block.
	ExpectInSize = 20 * 1024

	// EOF is the terminal EOF character - NOT guaranteed to be correct on all
	// systems nor if the program passed to expect connects to a different
	// platform
	EOF = "\004"
)

func init() {
	// No good way to discover EOF so have to take a best guess
	// Default (above) to control-D
	if runtime.GOOS == "windows" {
		EOF = "\032" // control-Z
	}
}

// Remember: Expect.Close() will not end the process
// you have to send it an EOF

type Expect struct {
	// *os.File is an anonymous field for the pty connected to the command.
	// This allows you to treat *Expect as a *os.File
	*os.File

	cmd    *exec.Cmd
	cmdIn  io.Reader // Internally I always read from cmdIn not File so cmdOut works
	cmdOut io.Writer

	timeout time.Duration

	// Expect reads into here. On a successful match all data up the end of the
	// match is deleted shrinking the buffer.
	// See also Clear() and BufStr()
	Buffer *bytes.Buffer

	// expertReader reads from Cmd and sends to Expect over this Chan
	bytesIn chan byteIn

	// On EOF being read from Cmd this is set (and ExpectReader is ended)
	Eof bool

	// Result is filled in asynchronously after the cmd exits
	Result ExpectWaitResult

	// Cancellation context: @cancel causes internal and os/exec cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

type ExpectWaitResult struct {
	ProcessState *os.ProcessState
	Error        error
}

// debugf logs only if Debug is true
func debugf(format string, args ...interface{}) {
	if Debug {
		log.Printf(format, args...)
	}
}

// NewExpect starts prog, passing any given args, in its own pty.
// Note that in order to be non-blocking while reading from the pty this sets
// the non-blocking flag and looks for EAGAIN on reads failing.  This has only
// been tested on Linux systems.
// On prog exiting or being killed Result is filled in shortly after.
func NewExpect(prog string, arg ...string) (*Expect, error) {
	return newExpectCommon(context.Background(), true, prog, arg...)
}

// NewExpectProc is similar to NewExpect except the created cmd is returned.
// It is expected that the cmd will exit normally otherwise it is left to
// the caller to kill it. This can be achieved by canceling the context @ctx.
// In the event of an error starting the cmd it will be killed but not reaped.
// However the cmd ends it is important that the caller reap the process
// by calling cmd.Process.Wait() otherwise it can use up a process slot in
// the operating system.
// Note that Result is not filled in.
func NewExpectProc(ctx context.Context, prog string, arg ...string) (*Expect, *exec.Cmd, error) {
	exp, err := newExpectCommon(ctx, false, prog, arg...)
	if err != nil {
		return nil, nil, err
	}
	return exp, exp.cmd, err
}

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

	if exp.File, err = pty.Start(exp.cmd); err != nil {
		return nil, err
	}

	// make the pty non blocking so when I read from it I dont jam up
	if err = syscall.SetNonblock(int(exp.File.Fd()), true); err != nil {
		return nil, err
	}
	exp.SetCmdOut(nil)

	go exp.expectReader()

	if reap {
		go exp.expectReaper()
	}

	return exp, nil
}

// expectReaper reaps the process if it ends for any reason and saves the
// Wait() result
func (exp *Expect) expectReaper() {
	exp.Result.ProcessState, exp.Result.Error = exp.cmd.Process.Wait()
}

// SetCmdOut if a non-nil io.Writer is passed it will be sent a copy of everything
// read by Expect() from the pty.
// Note that if you bypass expect and read directly from the *Expect this is
// will not be used
func (exp *Expect) SetCmdOut(cmdOut io.Writer) {
	if cmdOut != nil {
		exp.cmdOut = cmdOut
		exp.cmdIn = io.TeeReader(exp.File, cmdOut)
	} else {
		exp.cmdIn = exp.File
		exp.cmdOut = nil
	}
}

// SetTimeout sets the timeout for future calls to Expect().
// The default value is zero which cause Expect() to wait forever.
func (exp *Expect) SetTimeout(timeout time.Duration) {
	exp.timeout = timeout
}

// SetTimeoutSecs is a convenience wrapper around SetTimeout
func (exp *Expect) SetTimeoutSecs(timeout int) {
	exp.SetTimeout(time.Duration(timeout) * time.Second)
}

// Expect keeps reading input till either a timeout occurs (if set), one of the
// strings/regexps passed matches the input, end of input occurs or an error.
// If a string/regexp match occurs the index of the successful argument and the matching bytes
// are returned. Otherwise an error value and error are returned.
// Note: on EOF the return value will be NotFound and the error will be nil as
// EOF is not considered an error. This is the only time those values will be returned.
// See also Expecti()
func (exp *Expect) Expect(reOrStrs ...interface{}) (int, []byte, error) {
	// Check the args
	for n, reOrStr := range reOrStrs {
		switch reOrStr.(type) {
		case string:
			continue
		case *regexp.Regexp:
			continue
		default:
			debugf("Expect non string/regexp passed as arg %d", n)
			return NotStringOrRexgexp, nil, ENotStringOrRexgexp
		}
	}

	if exp.Eof {
		debugf("already at EOF")
		return NotFound, nil, nil
	}

	timedOut := make(<-chan time.Time)

	if exp.timeout != 0 {
		timedOut = time.After(exp.timeout)
	}

	if len(exp.Buffer.Bytes()) > 0 {
		// Fake byte-less input to get the for/select loop to process any pending buffered
		// input
		exp.bytesIn <- byteIn{}
	}

	for {
		select {
		case <-exp.ctx.Done():
			debugf("Expect got canceled: %s", exp.ctx.Err())
			return TimedOut, nil, exp.ctx.Err()
		case <-timedOut:
			debugf("Expect timedOut")
			return TimedOut, nil, ETimedOut
		case boe, ok := <-exp.bytesIn:
			if !ok {
				debugf("Expect read error")
				exp.Eof = true
				return NotFound, nil, EReadError
			}

			if boe.isEOF {
				debugf("Expect eof")
				exp.Eof = true
				return NotFound, nil, nil
			}

			if boe.isByte {
				b := boe.b
				debugf("Expect got new byte %c", b)
				if err := exp.Buffer.WriteByte(b); err != nil {
					debugf("Expect failed to add to buffer: %s", err)
					return NotFound, nil, EReadError
				}
			}

			bufBytes := exp.Buffer.Bytes()
			debugf("Expect buffer now:<<%s>>", string(bufBytes))
			debugf("Expect check for regexps")
			for n, reOrStr := range reOrStrs {
				var start, end int
				switch rs := reOrStr.(type) {
				case string:
					debugf("string passed: %s", rs)
					start = bytes.Index(bufBytes, []byte(rs))
					if start < 0 {
						continue
					}
					end = start + len(rs)
					debugf("string found")
				case *regexp.Regexp:
					debugf("re passed: %s", rs)
					loc := rs.FindIndex(bufBytes)
					if loc == nil {
						continue
					}
					start, end = loc[0], loc[1]
					debugf("re found")
				}

				// dont just assign a slice as I'm about to change the contents
				// of bytes and the slice will end up referencing the new data
				//found := bytes[start:end]
				found := make([]byte, end-start)
				copy(found, bufBytes[start:end])
				debugf("Expect found %s (start %d, end %d)", string(found), start, end)

				debugf("Expect reset buffer to the remaining input following the match")
				debugf("Expect buffer before reset:<<%s>>", string(exp.Buffer.Bytes()))
				newBuf := bufBytes[end:]
				debugf("Expect remaining:<<%s>>", string(newBuf))
				exp.Buffer.Reset()
				exp.Buffer.Write(newBuf)
				debugf("Expect buffer after reset:<<%s>>", string(exp.Buffer.Bytes()))

				return n, found, nil
			}
		}
	}

	// I can never get here...
	return NotFound, nil, nil
}

// byteOrEof is used between Expect and readToChan.
// If isEOF is false and isByte is false then there is no input. This is used
// to get Expect() to process left over buffer input.
type byteIn struct {
	isEOF  bool
	isByte bool
	b      byte
}

// expectReader reads from the pty and sends either a byte or eof to Expect.
func (exp *Expect) expectReader() {
	debugf("expectReader starting")
	buf := make([]byte, 1)
	for {
		select {
		case <-exp.ctx.Done():
			debugf("expectReader ending")
			return
		default:
			n, err := exp.cmdIn.Read(buf)
			debugf("expectReader read %d, %c, %v", n, buf[0], err)
			if err != nil {
				if unixIsEAGAIN(err) {
					debugf("expectReader EAGAIN")
					time.Sleep(100 * time.Millisecond) // reduce busy looping
					continue
				}
				debugf("expectReader ending read error")
				exp.bytesIn <- byteIn{isEOF: true}
				return
			}
			if n == 0 {
				// Not EAGAIN but no input
				debugf("expectReader ending nothing read")
				exp.bytesIn <- byteIn{isEOF: true}
				return
			}
			if n < 0 {
				continue
			}
			exp.bytesIn <- byteIn{isByte: true, b: buf[0]}
		}
	}
}

// Clear out any unprocessed input
func (exp *Expect) Clear() {
	exp.Buffer.Reset()
}

// BufStr is the buffer of expect read data as a string
func (exp *Expect) BufStr() string {
	return string(exp.Buffer.Bytes())
}

// Kill the command. Using an Expect after a Kill is undefined
func (exp *Expect) Kill() {
	exp.Buffer.Reset()
	exp.File.Close()
	exp.cancel() // calls os.Process.Kill() on exp.cmd
}

// Original expect compatibility

// Spawn is a wrapper to NewExpect(), for compatibility with the original expect
func Spawn(prog string, arg ...string) (*Expect, error) {
	return NewExpect(prog, arg...)
}

// Send sends the string to the process, for compatibility with the original expect
func (exp *Expect) Send(s string) (int, error) {
	return exp.Write([]byte(s))
}

// Sends string rune by rune with a delay before each.
// Note: the return is the number of bytes sent not the number of runes sent
func (exp *Expect) SendSlow(delay time.Duration, s string) (int, error) {
	for _, rune := range s {
		time.Sleep(delay)
		bytes := []byte(string(rune))
		n, err := exp.Write(bytes)
		if err != nil {
			return n, err
		}
	}
	return len([]byte(s)), nil
}

// SendL is a convenience wrapper around Send(), adding linebreaks around each of the @lines.
func (exp *Expect) SendL(lines ...string) error {
	for _, line := range lines {
		if _, err := exp.Send(line + "\r"); err != nil {
			return err
		}
	}
	return nil
}

// Expecti is a convenience wrapper around Expect() that only returns the index
// and not the found bytes or error. This is close to the original expect()
func (exp *Expect) Expecti(reOrStrs ...interface{}) int {
	i, _, _ := exp.Expect(reOrStrs...)
	return i
}

// Expectp is a convenience wrapper around Expect() that panics on no match
// (so will panic on eof)
func (exp *Expect) Expectp(reOrStrs ...interface{}) int {
	i, _, _ := exp.Expect(reOrStrs...)
	if i < 0 {
		panic(fmt.Sprintf("%s", reOrStrs...))
	}
	return i
}

// LogUser true asks for all input read by Expect() to be copied to stdout.
// LogUser false (default) turns it off.
// For compatibility with the original expect
func (exp *Expect) LogUser(on bool) {
	if on {
		exp.SetCmdOut(os.Stdout)
	} else {
		exp.SetCmdOut(nil)
	}
}

// Copied from the non-exported func in src/crypto/rand/eagain.go

func unixIsEAGAIN(err error) bool {
	if pe, ok := err.(*os.PathError); ok {
		if errno, ok := pe.Err.(syscall.Errno); ok && errno == syscall.EAGAIN {
			return true
		}
	}
	return false
}
