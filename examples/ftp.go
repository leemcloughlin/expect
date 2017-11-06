// ftp example thanks to github.com/ThomasRooney/gexpect
package main

import (
	"log"

	"github.com/grrtrr/expect"
)

func main() {
	exp, err := expect.Spawn("ftp", "openbsd.cs.toronto.edu")
	if err != nil {
		log.Fatalf("failed to spawn OpenBSD ftp: %s", err)
	}
	exp.LogUser(true)

	exp.Expect("Name")
	exp.SendL("anonymous")
	exp.Expect("Password")
	exp.SendL("expect@sourceforge.net")
	exp.Expect("ftp> ")
	exp.SendL("cd /pub/OpenBSD/5.8")
	exp.Expect("ftp> ")
	exp.SendL("dir")
	exp.Expect("ftp> ")
	exp.SendL("prompt")
	exp.Expect("ftp> ")
	exp.SendL("pwd")
	exp.Expect("ftp> ")
}
