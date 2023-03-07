package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/federicotdn/pimacs/elisp"
	"os"
	"strings"
)

func repl() {
	interpreter := elisp.NewInterpreter()

	for {
		reader := bufio.NewReader(os.Stdin)
		eval := true

		fmt.Print("> ")
		source, _ := reader.ReadString('\n')
		source = strings.TrimRight(source, "\r\n")

		if source == "" {
			break
		} else if source[0] == '|' {
			eval = false
			source = source[1:]
			fmt.Println("[input will not be evaluated]")
		}

		var printed string
		var err error

		if eval {
			printed, err = interpreter.ReadEvalPrin1(source)
		} else {
			printed, err = interpreter.ReadPrin1(source)
		}

		if err != nil {
			fmt.Println("repl error:", err)
		} else {
			fmt.Println(printed)
		}
	}
}

func load(filename string) {
	interpreter := elisp.NewInterpreter()

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	source := string(data)
	printed, err := interpreter.ReadEvalPrin1(source)

	if err != nil {
		fmt.Println("load error:", err)
		os.Exit(1)
	}

	fmt.Println(printed)
}

func main() {
	const usage = "load Emacs Lisp FILE using the load function"

	var filename string
	flag.StringVar(&filename, "load", "", usage)
	flag.StringVar(&filename, "l", "", usage+" (shorthand)")
	flag.Parse()

	if filename != "" {
		load(filename)
	} else {
		repl()
	}
}
