package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mndrix/getkey"
)

func main() {
	var d bool
	var n int
	var p string

	flag.BoolVar(&d, "d", false, "enable debug mode")
	flag.IntVar(&n, "n", 1, "number of key presses to read")
	flag.StringVar(&p, "p", "", "prompt before awaiting a key")
	flag.Parse()
	getkey.SetDebug(d)

	pledge()
	for i := 0; i < n; i++ {
		if p != "" {
			os.Stdout.Write([]byte(p))
		}
		c, err := getkey.GetKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			os.Exit(1)
		}
		if p != "" {
			fmt.Print("\n")
		}
		fmt.Println(c)

	}
	return
}
