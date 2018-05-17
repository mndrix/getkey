package getkey // import "github.com/mndrix/getkey"
import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

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
		debugf("matches = %v", matches)
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
			debugf("args = %v", args)

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
