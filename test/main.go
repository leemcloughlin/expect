/*
File summary: Test app to produce just used by expect running go test
Package: expect
Author: Lee McLoughlin

Copyright (C) 2016 LMMR Tech Ltd


NOTE: Most of these tests use the Linux od command. Without the same version of
od they will fail
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("Args passed:", os.Args)

	if len(os.Args) == 2 && os.Args[1] == "nothing" {
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter test name: ")
		name, _ := reader.ReadString('\n')
		name = strings.Trim(name, "\n")
		switch name {
		case "":
			fmt.Println("No input goodbye")
			os.Exit(0)
		case "0":
			fmt.Println("Goodbye")
			os.Exit(0)
		case "1":
			fmt.Println("Welcome to the first test")
		case "2":
			fmt.Println("Welcome to the second test")
			fmt.Println("Two lines of output!")
		case "3":
			fmt.Println("Abcdef")
			fmt.Printf("gh")
			time.Sleep(5 * time.Second)
			fmt.Println("ijk")
		case "4":
			fmt.Println("世界")
		case "HELLO":
			fmt.Println("I saw hello")
		default:
			fmt.Printf("unknown test <<%s>>\n", name)
		}
	}
}
