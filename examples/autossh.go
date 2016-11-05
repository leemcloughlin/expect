//
// Automates the ssh key exchange for password-less login on a remote system.
//
// Original taken from http://www.techpaste.com/2013/04/28/shell-script-automate-ssh-key-transfer-hosts-linux/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/leemcloughlin/expect"
)

// Name of key (in user's $HOME/.ssh directory) to exchange
const (
	keyName = "id_rsa.pub"
)

func main() {
	var remoteUser = flag.String("l", "", "Remote user to use (defaults to local user)")
	var password = flag.String("p", "", "Password for remote user")
	var verbose = flag.Bool("v", true, "Log user interaction to stdout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <remoteHost>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 || *password == "" {
		flag.Usage()
		os.Exit(1)
	}
	host := flag.Arg(0)

	user, err := user.Current()
	if err != nil {
		log.Fatalf("failed to look up current user: %s", err)
	}

	if *remoteUser == "" {
		*remoteUser = user.Username
	}

	keyPath := path.Join(user.HomeDir, ".ssh", keyName)
	if _, err := os.Stat(keyPath); err != nil {
		log.Fatalf("Can not use ssh key %s: %s", keyPath, err)
	}

	// First remove existing entries for @host from ~/.ssh/known_hosts
	exp, err := expect.Spawn("ssh-keygen", "-R", host)
	if err != nil {
		log.Fatalf("failed to remove existing hostkeys for %s: %s", host, err)
	}
	exp.LogUser(*verbose)

	_, _, err = exp.Expect("known_hosts: No such file or directory", "Original contents retained")
	if err != nil {
		log.Fatalf("unexpected interaction removing %s from known_hosts: %s", host, err)
	}
	exp.Kill()

	exp, err = expect.Spawn("ssh-copy-id", "-i", keyPath, fmt.Sprintf("%s@%s", *remoteUser, host))
	if err != nil {
		log.Fatalf("filed to run ssh-copy-id: %s", err)
	}
	exp.LogUser(*verbose)

	// We have removed the known_hosts key, so it will prompt to add it back again
	exp.Expect("(yes/no)?")
	exp.SendL("yes")

	for {
		i, _, err := exp.Expect(
			"assword:",                    // standard password prompt
			"Permission denied",           // password received, but incorrect
			"already exist on the remote", // key already exists
			"Now try",                     // issued after transfer
		)
		if err != nil {
			log.Fatalf("expect error: %s", err)
		} else if i == 3 {
			log.Printf("Successfully transferred ssh key for %s@%s", *remoteUser, host)
			break
		} else if i == 2 {
			log.Printf("Key for %s already exists on %s", *remoteUser, host)
			break
		} else if i == 1 {
			log.Fatalf("Wrong password (permission denied) for %s@%s", *remoteUser, host)
		} else if i == 0 {
			exp.SendL(*password)
		} else {
			log.Fatalf("unexpected response - try -v to debug")
		}
	}
}
