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
	"github.com/pkg/errors"
)

// Terminal represents a terminal which has been prepared for fetching
// single keystrokes.
type Terminal struct {
	// term is the way we interact with the underlying terminal device
	term *term.Term

	// buf is a buffer into which raw terminal escape sequences are
	// read
	buf []byte
}

// Prepare opens and configures the terminal so that a single
// character can be fetched with high precision.  When finished
// reading characters from the terminal, invoke the Restore method.
func Prepare() (*Terminal, error) {
	var err error
	t := &Terminal{}
	t.term, err = term.Open("/dev/tty")
	if err != nil {
		return nil, errors.Wrap(err, "preparing terminal")
	}

	t.buf = make([]byte, 15)

	// TODO should probably use termcap or something
	term := os.Getenv("TERM")
	if term == "xterm" || strings.HasPrefix(term, "xterm-") {
		// TODO
		// if bash is capturing our stdout into a variable,
		// these escapes cause the variable to include escape codes
		// so comparing it against "Ctrl-n" (for example) always
		// returns false.
		// Maybe sending it to stderr would work?
		// Maybe sending it directly to /dev/tty would work?

		//fmt.Print("\x1b[>4;2m")     // xterm: set modifyOtherKeys=2
		//fmt.Print("x")              // xterm eats this character
		// defer fmt.Print("\x1b[>4m") // xterm: restore modifyOtherKeys
	}

	return t, nil
}

// read returns a sequence of raw bytes read from the terminal.
func (t *Terminal) read() ([]byte, error) {
	term.RawMode(t.term)
	defer t.term.Restore()
	numRead, err := t.term.Read(t.buf)
	if err != nil {
		return nil, errors.Wrap(err, "reading")
	}

	return t.buf[0:numRead], nil
}

// Restore returns the terminal to its original state.  It should be
// called when you're done reading single characters from the
// terminal.
func (t *Terminal) Restore() error {
	err := t.term.Close()
	if err != nil {
		return errors.Wrap(err, "restoring")
	}
	return nil
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

	t, err := Prepare()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}
	defer t.Restore()

	for i := 0; i < *n; i++ {
		if *p != "" {
			os.Stdout.Write([]byte(*p))
		}
		c, err := t.read()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			os.Exit(1)
		}
		if *p != "" {
			fmt.Print("\n")
		}

		// TODO move all this decoding into `func (t *Terminal) GetCh() string`
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
