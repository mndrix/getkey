package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/mndrix/term"
	"github.com/pkg/errors"
)

var terminal *term.Term
var envTerm string
var buf []byte
var mux sync.Mutex

var isDebug bool

func init() {
	var err error
	terminal, err = term.Open("/dev/tty")
	if err != nil {
		panic(err)
	}

	envTerm = os.Getenv("TERM")
	buf = make([]byte, 15)
}

// should only be called while holding mutex
func prepare() {
	// TODO should probably use termcap or something
	if envTerm == "xterm" || strings.HasPrefix(envTerm, "xterm-") {
		// TODO
		// if bash is capturing our stdout into a variable,
		// these escapes cause the variable to include escape codes
		// so comparing it against "Ctrl-n" (for example) always
		// returns false.
		// Maybe sending it to stderr would work?
		// Maybe sending it directly to /dev/tty would work?

		//fmt.Print("\x1b[>4;2m")     // xterm: set modifyOtherKeys=2
		//fmt.Print("x")              // xterm eats this character
	}
}

// should only be called while holding mutex
func restore() {
	if envTerm == "xterm" || strings.HasPrefix(envTerm, "xterm-") {
		// fmt.Print("\x1b[>4m") // xterm: restore modifyOtherKeys
	}
}

// returns a sequence of raw bytes read from the terminal.
// should only be called while holding mutex.
func read() ([]byte, error) {
	mux.Lock()
	defer mux.Unlock()

	term.RawMode(terminal)
	defer terminal.Restore()

	prepare()
	defer restore()

	numRead, err := terminal.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "reading")
	}

	return buf[0:numRead], nil
}

// GetKey waits for the user to press a key and then returns a string
// representing what was pressed.  Alphanumeric characters and
// punctuation are represented as themselves.  Modifier keys are
// represented as Alt-, Ctrl-, Esc-, and Shift- prefixes (in that
// order) on the base key.  For example, holding down Control and Alt
// while pressing the L key produces "Alt-Ctrl-L".
func GetKey() (string, error) {
	raw, err := read()
	if err != nil {
		return "", errors.Wrap(err, "GetKey")
	}
	return decode(raw)
}

func main() {
	flag.BoolVar(&isDebug, "d", false, "enable debug mode")
	var n = flag.Int("n", 1, "number of key presses to read")
	var p = flag.String("p", "", "prompt before awaiting a key")
	flag.Parse()

	for i := 0; i < *n; i++ {
		if *p != "" {
			os.Stdout.Write([]byte(*p))
		}
		c, err := GetKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			os.Exit(1)
		}
		if *p != "" {
			fmt.Print("\n")
		}
		fmt.Println(c)

	}
	return
}

func debugf(format string, args ...interface{}) {
	if isDebug {
		fmt.Printf("DEBUG: "+format+"\n", args...)
	}
}

// decode a single sequence of raw bytes from the terminal
func decode(c []byte) (string, error) {
	s := string(c)
	debugf("raw: %v %q", c, s)
	if len(c) == 1 {
		debugf("single character")
		r := rune(c[0])
		if name := runeName(r); name != "" {
			return name, nil
		}
		if r == 0 {
			return "Ctrl-" + runeName(' '), nil
		}
		if r <= 26 {
			return "Ctrl-" + runeName(rune(r+96)), nil
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
					return modifier + runeName(r), nil
				}
			case "H":
				switch len(args) {
				case 0:
					return "Home", nil
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
	}
	return s, nil
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
