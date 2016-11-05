// Classical !dlrow olleh example
package main

import (
	"fmt"
	"os"

	"github.com/leemcloughlin/expect"
)

func main() {
	exp, err := expect.NewExpect("rev")
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
	exp.Kill()
}
