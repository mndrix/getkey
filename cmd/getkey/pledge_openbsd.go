package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func pledge() {
	// restrict privileges on OpenBSD
	err := unix.PledgePromises("stdio tty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}
}
