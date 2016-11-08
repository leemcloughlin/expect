/*
File summary: go test
Package: expect
Author: Lee McLoughlin

Copyright (C) 2016 LMMR Tech Ltd

Note: first build the test app:
	cd test
	go build
*/

package expect

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	//prog = "/usr/bin/od"
	prog = "test/test"
)

func TestMain(m *testing.M) {
	flag.BoolVar(&Debug, "debug", false, "debugging")
	flag.Parse()

	if Debug {
		fmt.Fprintf(os.Stderr, "Debugging ON\n")
	}

	_, err := os.Stat(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot find %s: have you run \"go build\" in the test sub directory?\n", prog)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	fullname := runtime.FuncForPC(pc).Name()
	// convert lmmrtech.com/lee/expect.Test_NewExpect to Test_NewExpect
	lastDot := strings.LastIndex(fullname, ".")
	if lastDot <= 0 {
		return fullname
	}
	return fullname[lastDot+1:]
}

func checkResultStr(t *testing.T, pat string, i int, n int, found []byte, err error) {
	debugf("n %d, found %s, err %s", n, string(found), err)
	if n == i {
		if string(found) != pat {
			t.Errorf("expected %s but got %s", pat, string(found))
		} else {
			t.Logf("found expected str: %s", string(found))
		}
	} else {
		t.Errorf("did not find expected str: %s", err)
	}
}

func checkResultRe(t *testing.T, pat string, re *regexp.Regexp, i int, n int, found []byte, err error) {
	debugf("n %d, found %s, err %s", n, string(found), err)
	if n == i {
		if string(found) != pat {
			t.Errorf("expected %s but got %s", pat, string(found))
		} else {
			t.Logf("found expected RE: %s", string(found))
		}
	} else {
		t.Errorf("did not find expected RE: %s", err)
	}
}

func showWaitResult(t *testing.T, exp *Expect) {
	valid := false
	// Wait a little bit for Result to be filled in
	time.Sleep(time.Millisecond * 10)
	// Try a few fimes to see if its valid (maybe I should have used a channel
	// but checking for all the errors is painful for code I expect to be used
	// so rarely)
	for i := 1; i <= 3; i++ {
		if exp.Result.ProcessState != nil || exp.Result.Error != nil {
			valid = true
			break
		}
		t.Logf("pause %d for Wait() result", i)
		time.Sleep(time.Second)
	}
	if !valid {
		t.Logf("Wait() result never went valid")
		return
	}
	if exp.Result.Error != nil {
		t.Logf("Wait() result: %s, %s", exp.Result.ProcessState, exp.Result.Error)
	} else {
		t.Logf("Wait() result: %s", exp.Result.ProcessState)
	}
}

func Test_NewExpect(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
		return
	}
	t.Log("OK killing processes")
	exp.Kill()
	showWaitResult(t, exp)
}

func Test_NewExpectProc(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, cmd, err := NewExpectProc(context.TODO(), prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
		return
	}
	t.Log("OK killing processes")
	exp.Kill()
	ps, err := cmd.Process.Wait()
	if err != nil {
		t.Errorf("Failed to Wait() %s", err)
		return
	}
	t.Logf("cmd Wait() result: %s", ps)
}

func Test_Spawn(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := Spawn(prog)
	if err != nil {
		t.Errorf("Spawn failed %s", err)
		return
	}
	t.Log("OK killing processes")
	exp.Kill()
	showWaitResult(t, exp)
}

func Test_LogUser(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	fmt.Fprintf(os.Stderr, "%s testing:\nAt some point you should see:\n%s",
		funcName(),
		`
0
Args passed: [test/test]
Enter test name: Goodbye

(It ends with Goodbye)
This may appear before the log messages as they are buffered

`)

	t.Logf("starting %s", prog)
	exp, err := Spawn(prog)
	if err != nil {
		t.Errorf("Spawn failed %s", err)
		return
	}
	t.Log("enabling LogUser - you should now see the test program output on stdout")
	exp.LogUser(true)
	exp.SetTimeoutSecs(5)
	exp.Send("0\r")
	exp.Expect()

	t.Log("killing process")
	exp.Kill()
	showWaitResult(t, exp)
}

func Test_ExpectCmdOutAndClear(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	buf := new(bytes.Buffer)
	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(5) // Shouldn't happen

	// As Expect to save output here
	exp.SetCmdOut(buf)

	exp.Send("1\r")
	exp.Send(EOF)

	// This will grab everything (till eof) and will copy it to buf
	exp.Expect()
	expected := "Welcome to the first test"
	bufs := buf.String()
	if !strings.Contains(bufs, expected) {
		t.Errorf("buf wrong contains <<%s>> not <<%s>>", bufs, expected)
		for i, c := range bufs {
			t.Errorf("c[%d] %c %d", i, c, int(c))
		}
		// t.Errorf("buf wrong len %d not %d", len(bufs), len(expected))
	} else {
		t.Log("buf contains expected string")
	}

	t.Logf("clearing buffer")
	exp.Clear()
	bufs = exp.BufStr()
	if bufs != "" {
		t.Errorf("buf wrong contains <<%s>> should be empty", bufs, expected)
		for i, c := range bufs {
			t.Errorf("c[%d] %c %d", i, c, int(c))
		}
		// t.Errorf("buf wrong len %d not %d", len(bufs), len(expected))
	} else {
		t.Log("buf is empty as expected")
	}
	showWaitResult(t, exp)
}

func Test_ExpectMatchTimeout(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}

	t.Log("setting 1 sec timer and sending nothing (including no EOF) so will hang forcing timeout")
	exp.SetTimeoutSecs(1)

	n, found, err := exp.Expect("no way")
	if n == 0 {
		t.Errorf("found something when I shouldn't: %s", string(found))
	} else if err == ETimedOut {
		t.Logf("timed out as expected")
	} else {
		t.Logf("found unexpected error: %s", err)
	}

	exp.Kill()
	showWaitResult(t, exp)
}

func Test_ExpectMatch1s(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 1\\r + eof")
	exp.Send("1\r")

	pat := "Welcome to the first test"
	n, found, err := exp.Expect(pat)
	checkResultStr(t, pat, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatch1re(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 1\\r + eof")
	exp.Send("1\r")

	pat := "Welcome to the first test"
	re := regexp.MustCompile(pat)
	n, found, err := exp.Expect(re)
	checkResultRe(t, pat, re, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectiMatch1(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 1\\r + eof")
	exp.Send("1\r")

	pat := "Welcome to the first test"
	n := exp.Expecti(pat)
	if n != 0 {
		t.Errorf("expected 0 got %d", n)
	} else {
		t.Log("OK")
	}

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatch2re(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 2\\r + eof")
	exp.Send("2\r")

	//pat := "Welcome to the second test"
	//pat2 := "Two lines of output!"
	pat := "Welcome "
	pat2 := "to the second test"

	re := regexp.MustCompile(pat)
	re2 := regexp.MustCompile(pat2)

	n, found, err := exp.Expect(re, re2)
	checkResultRe(t, pat, re, 0, n, found, err)

	n, found, err = exp.Expect(re, re2)
	checkResultRe(t, pat2, re2, 1, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatch2s(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 2\\r + eof")
	exp.Send("2\r")

	//pat := "Welcome to the second test"
	//pat2 := "Two lines of output!"
	pat := "Welcome "
	pat2 := "to the second test"

	n, found, err := exp.Expect(pat, pat2)
	debugf("n %d, found %s, err %s", n, string(found), err)
	checkResultStr(t, pat, 0, n, found, err)

	n, found, err = exp.Expect(pat, pat2)
	checkResultStr(t, pat2, 1, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatch2reEOI(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 2\\r + eof")
	exp.Send("2\r")

	//pat := "Welcome to the second test"
	//pat2 := "Two lines of output!"
	pat := "test\r\n"
	pat2 := "Two lines"

	re := regexp.MustCompile(pat)
	re2 := regexp.MustCompile(pat2)

	n, found, err := exp.Expect(re, re2)
	debugf("n %d, found %s, err %s", n, string(found), err)
	checkResultRe(t, pat, re, 0, n, found, err)

	n, found, err = exp.Expect(re, re2)
	checkResultRe(t, pat2, re2, 1, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatch2sEOI(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	t.Log("sending 2\\r + eof")
	exp.Send("2\r")

	//pat := "Welcome to the second test"
	//pat2 := "Two lines of output!"
	pat := "test\r\n"
	pat2 := "Two lines"

	n, found, err := exp.Expect(pat, pat2)
	debugf("n %d, found %s, err %s", n, string(found), err)
	checkResultStr(t, pat, 0, n, found, err)

	n, found, err = exp.Expect(pat, pat2)
	checkResultStr(t, pat2, 1, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatchFind2nd(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog)
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10) // Shouldn't happen

	exp.Expect("Enter test name:")

	t.Log("sending 2\\r + eof")
	exp.Send("2\r")

	pat := "DONT FIND THIS"
	pat2 := "Two lines of output!"

	n, found, err := exp.Expect(pat, pat2)
	debugf("n %d, found %s, err %s", n, string(found), err)
	checkResultStr(t, pat2, 1, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatchSplitRe(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog, "somearg")
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10)

	exp.Expect("Enter test name:")

	t.Log("sending 3\\r + eof")
	exp.Send("3\r")

	t.Log("The test program will send back two lines with a delay part way through the 2nd")
	pat := "Abcdef"
	pat2 := "ghijk"

	re := regexp.MustCompile(pat)
	n, found, err := exp.Expect(re)
	checkResultRe(t, pat, re, 0, n, found, err)

	re = regexp.MustCompile(pat2)
	n, found, err = exp.Expect(re)
	checkResultRe(t, pat2, re, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectMatchSplitS(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog, "somearg")
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10)

	exp.Expect("Enter test name:")

	t.Log("sending 3\\r + eof")
	exp.Send("3\r")

	t.Log("The test program will send back two lines with a delay part way through the 2nd")
	pat := "Abcdef"
	pat2 := "ghijk"

	n, found, err := exp.Expect(pat)
	checkResultStr(t, pat, 0, n, found, err)

	n, found, err = exp.Expect(pat2)
	checkResultStr(t, pat2, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectSendDelay(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog, "somearg")
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10)

	exp.Expect("Enter test name:")

	// I'm sending two lines but the 2nd line I dont check its just to test
	// sending non-English runes
	// I also send HELLO but expect back: I saw hello
	t.Log("sending HELLO\\r世界\\r + eof")
	sent, _ := exp.SendSlow(time.Second, "HELLO\r世界\r")
	t.Logf("sent %d runes", sent)

	pat := "I saw hello"
	n, found, err := exp.Expect(pat)
	checkResultStr(t, pat, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectUTFs(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog, "somearg")
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10)

	exp.Expect("Enter test name:")

	t.Log("sending 4\\r + eof")
	exp.Send("4\r")

	pat := "世界"
	n, found, err := exp.Expect(pat)
	checkResultStr(t, pat, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}

func Test_ExpectUTFre(t *testing.T) {
	debugf("%s start", funcName())
	defer debugf("%s end", funcName())

	t.Logf("starting %s", prog)
	exp, err := NewExpect(prog, "somearg")
	if err != nil {
		t.Errorf("NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(10)

	exp.Expect("Enter test name:")

	t.Log("sending 4\\r + eof")
	exp.Send("4\r")

	pat := "世\\p{L}"
	re := regexp.MustCompile(pat)
	n, found, err := exp.Expect(re)
	checkResultRe(t, "世界", re, 0, n, found, err)

	exp.Send(EOF)
	showWaitResult(t, exp)
}
