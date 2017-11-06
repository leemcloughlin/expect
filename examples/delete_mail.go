// Delete all system mails of the current user, like 'yes d | mail'
// Please use with care :-)
package main

import (
	"log"

	"github.com/grrtrr/expect"
)

func main() {
	exp, err := expect.Spawn("mail")
	if err != nil {
		log.Fatalf("failed to spawn mail program: %s", err)
	}
	exp.LogUser(true)

	for {
		// '&' is the mail user prompt
		i, _, err := exp.Expect("No mail", "& ", "No applicable messages")
		if err != nil {
			log.Fatalf("failed to match expressed expressions: %s", err)
		}

		if i == 0 { // Nothing to delete
			break
		} else if i == 1 { // prompt
			exp.SendL("d")
		} else if i == 2 { // no applicable messages - everything deleted
			exp.SendL("q")
			break
		} else if i == expect.NotFound {
			log.Fatalf("program closed unexpectedly")
		}
	}
}
