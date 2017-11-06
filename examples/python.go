// Interact with the Python interpreter using go expect.
package main

import (
	"log"

	"github.com/grrtrr/expect"
)

func main() {
	exp, err := expect.Spawn("python")
	if err != nil {
		log.Fatalf("Failed to spawn python process: %s", err)
	}
	exp.LogUser(true)

	exp.Expect(">>>")
	exp.SendL("import time", "print 'It is now %s' % time.ctime(time.time())", "exit()")
	exp.Expect("exit()")
}
