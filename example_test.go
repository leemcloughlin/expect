/*
File summary: testable examples
Package: expect
Author: Lee McLoughlin

Copyright (C) 2016 LMMR Tech Ltd

*/

package expect

import (
	"fmt"
	"os"
	"testing"
)

func ExampleExpect() {
	// Run rev, the reverse text lines command, send it hello and
	// look for olleh
	exp, err := NewExpect("rev")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(5) // Shouldn't happen
	exp.Send("hello\r")

	i, found, err := exp.Expect("olleh")
	if i == 0 {
		fmt.Println("found", string(found))
	} else {
		fmt.Println("failed ", err)
	}
	exp.Send(EOF)
	// Output:
	// found olleh
}

/*
This example works and the code is small but I just really dislike
using panic's

func Test_Examplep(tst *testing.T) {
	// https://en.wikipedia.org/wiki/Expect
	fmt.Fprintln(os.Stderr, "Using expectp to panic on error - not a Go thing to do")
	defer func() {
		if r := recover(); r != nil {
			tst.Errorf("failed: %s", r)
		}
	}()
	//remote_server := "example.com"
	remote_server := "localhost"
	my_user_id := "lee"
	my_password := "lee"
	my_command := "ls"
	t, err := Spawn("telnet", remote_server)
	if err != nil {
		tst.Errorf("failed to telnet: %s", err)
		return
	}
	tst.Logf("telnet spawned ok")
	t.SetTimeoutSecs(5)
	t.Expectp("username:")
	// Send the username, and then wait for a password prompt.
	t.Send(my_user_id + "\r")
	t.Expectp("password:")
	// Send the password, and then wait for a shell prompt.
	t.Send(my_password + "\r")
	t.Expectp("%")
	t.Clear()
	// Send the prebuilt command, and then wait for another shell prompt.
	t.Send(my_command + "\r")
	t.Expectp("%")
	// Capture the results of the command into a variable. This can be displayed, or written to disk.
	results := t.BufStr()
	// Exit the telnet session, and wait for a special end-of-file character.
	t.Send("exit\r")
	t.Expectp() // read EOF
	fmt.Fprintln(os.Stderr, "results are:", results)
	tst.Log("OK")
}
*/

func ExampleExpect_ownEcho() {
	// Run rev, the reverse text lines command, send it hello and
	// look for hello echo'd back then olleh
	exp, err := NewExpect("rev")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewExpect failed %s", err)
	}
	exp.SetTimeoutSecs(5) // Shouldn't happen
	exp.Send("hello\r")

	// Remember terminals echo, by default, so I will get back the hello I
	// just sent before its reverse
	i, found, err := exp.Expect("hello", "olleh")
	if i == 0 {
		fmt.Println("found", string(found))
	} else {
		fmt.Println("failed ", err)
	}

	i, found, err = exp.Expect("hello", "olleh")
	if i == 1 {
		fmt.Println("found", string(found))
	} else {
		fmt.Println("failed ", err)
	}

	exp.Send(EOF)

	i, _, _ = exp.Expect()
	if i == NotFound {
		fmt.Println("found EOF")
	}
	// Output:
	// found hello
	// found olleh
	// found EOF
}
