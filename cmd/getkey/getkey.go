package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/mndrix/term"
)

func getch() []byte {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 15)
	numRead, err := t.Read(bytes)
	t.Restore()
	t.Close()
	if err != nil {
		return nil
	}
	return bytes[0:numRead]
}

func main() {
	var d = flag.Bool("d", false, "enable debug mode")
	var n = flag.Int("n", 1, "number of key presses to read")
	var p = flag.String("p", "", "prompt before awaiting a key")
	flag.Parse()
	debugf := func(format string, args ...interface{}) {
		if *d {
			fmt.Printf("DEBUG: "+format+"\n", args...)
		}
	}

	// see https://emacs.stackexchange.com/a/13957/ for great detail

	term := os.Getenv("TERM")
	if term == "xterm" || strings.HasPrefix(term, "xterm-") {
		fmt.Print("\x1b[>4;2m")     // xterm: set modifyOtherKeys=2
		fmt.Print("x")              // xterm eats this character
		defer fmt.Print("\x1b[>4m") // xterm: restore modifyOtherKeys
	}
	for i := 0; i < *n; i++ {
		if *p != "" {
			os.Stdout.Write([]byte(*p))
		}
		c := getch()
		if *p != "" {
			fmt.Print("\n")
		}
		s := string(c)
		debugf("raw: %v %q\n", c, s)
		if len(c) == 1 {
			debugf("single character")
			r := rune(c[0])
			if name := runeName(r); name != "" {
				fmt.Println(name)
				continue
			}
			if r == 0 {
				fmt.Println("Ctrl-" + runeName(' '))
				continue
			}
			if r <= 26 {
				fmt.Println("Ctrl-" + runeName(rune(r+96)))
				continue
			}
			debugf("fallthrough")
		}
		if strings.HasPrefix(s, "\x1b[") {
			s = strings.TrimPrefix(s, "\x1b[")
			rx := regexp.MustCompile(`^([0-9;]*)([^0-9])$`)
			matches := rx.FindStringSubmatch(s)
			fmt.Printf("matches = %v\n", matches)
			if matches != nil {
				argStrings := strings.Split(matches[1], ";")
				args := make([]int, 0, len(argStrings))
				for _, argString := range argStrings {
					if argString != "" {
						arg, err := strconv.Atoi(argString)
						if err != nil {
							panic(err)
						}
						args = append(args, arg)
					}
				}
				fmt.Printf("args = %v\n", args)

				// TODO matches[2] should select a function
				// TODO call that function with args
				// TODO like: switch matches[2] { case "~": tilde(args...) }

				switch matches[2] {
				case "~":
					if args[0] == 27 {
						r := rune(args[2])

						modifier := "?-"
						switch args[1] {
						case 2:
							modifier = ""
							if !unicode.IsPrint(r) {
								modifier += "Shift-"
							}
						case 3:
							modifier = "Alt-"
						case 4:
							modifier = "Alt-"
							if !unicode.IsPrint(r) {
								modifier += "Shift-"
							}
						case 5:
							modifier = "Ctrl-"
						case 6:
							modifier = "Ctrl-"
							if !unicode.IsPrint(r) {
								modifier += "Shift-"
							}
						case 7:
							modifier = "Ctrl-Alt-"
						case 8:
							modifier = "Ctrl-Alt-"
							if !unicode.IsPrint(r) {
								modifier += "Shift-"
							}
						}
						fmt.Printf("%s%s\n", modifier, runeName(r))
						continue
					}
				case "H":
					switch len(args) {
					case 0:
						fmt.Println("Home")
						continue
					case 2:
						// "1;2H" is Shift-Home
						// "1;3H" is Alt-Home
						// "1;5H" is Ctrl-Home
					}
				}
			}
			/*
				} else if s == "2~" {
					fmt.Println("Insert")
					continue
				} else if s == "F" {
					fmt.Println("End")
					continue
				} else if s == "Z" {
					fmt.Println("Shift-Tab")
					continue
				} else if false {
					// regexp: ^([0-9;]*)([^0-9])$

					// "2~"   is Insert
					// "2;3~" is Alt-Insert
					// "2;5~" is Ctrl-Insert

					// "3;5~" is Ctrl-Delete
					// "3;2~" is Shift-Delete

					// "1;2F" is Shift-End
					// "1;3F" is Alt-End
					// "1;5F" is Ctrl-End

				}
			*/
		} else {
			fmt.Printf("%s\n", s)
		}
	}
	return
}

func runeName(r rune) string {
	switch r {
	case 8:
		return "Backspace"
	case 9:
		return "Tab"
	case 13:
		return "Enter"
	case 27:
		return "Escape"
	case 32:
		return "Space"
	case 127:
		return "Delete"
	}
	if r > 31 && unicode.IsPrint(r) {
		return fmt.Sprintf("%c", r)
	}
	return ""
}
