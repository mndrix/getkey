package getkey // import "github.com/mndrix/getkey"
import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var csiRx = regexp.MustCompile(`^([0-9;]*)([^0-9])$`)

// decode a single sequence of raw bytes from the terminal
func decode(c []byte) (string, error) {
	s := string(c)
	debugf("raw: %v %q", c, s)

	if len(c) == 1 {
		debugf("single byte")
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

	// parse CSI escapes
	if strings.HasPrefix(s, "\x1b[") {
		if name, err := parseCsi(s); err == nil {
			return name, nil
		} else {
			return s, nil
		}
	}

	// parse function key escapes
	if strings.HasPrefix(s, "\x1bO") {
		if name, err := parseFn(s); err == nil {
			return name, nil
		} else {
			return s, nil
		}
	}

	return s, nil
}

func parseCsi(s string) (string, error) {
	debugf("CSI")
	s = strings.TrimPrefix(s, "\x1b[")
	matches := csiRx.FindStringSubmatch(s)
	debugf("matches = %v", matches)
	if matches == nil {
		return s, nil
	}

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

	switch matches[2] {
	case "~":
		return csiTilde(args)
	case "h":
		return csih(args)
	case "A":
		return "Up", nil
	case "B":
		return "Down", nil
	case "C":
		return "Right", nil
	case "D":
		return "Left", nil
	case "F": // xterm sends this
		return "End", nil
	case "H": // xterm sends this
		return "Home", nil
	case "P": // st sends this
		return "Backspace", nil
	}
	return s, nil
}

func parseFn(s string) (string, error) {
	debugf("Fn")
	s = strings.TrimPrefix(s, "\x1bO")
	if s == "" {
		return s, errors.New("Fn escape missing")
	}

	switch s {
	case "P":
		return "F1", nil
	case "Q":
		return "F2", nil
	case "R":
		return "F3", nil
	case "S":
		return "F4", nil
	}
	return s, nil
}

func csiTilde(args []int) (string, error) {
	debugf("CSI ~")
	switch args[0] {
	case 1:
		return "Home", nil
	case 2:
		return "Insert", nil
	case 3:
		return "Delete", nil
	case 4:
		return "End", nil
	case 5:
		return "PageUp", nil
	case 6:
		return "PageDown", nil
	case 15:
		return "F5", nil
	case 17:
		return "F6", nil
	case 18:
		return "F7", nil
	case 19:
		return "F8", nil
	case 20:
		return "F9", nil
	case 21:
		return "F10", nil
	case 23:
		return "F11", nil
	case 24:
		return "F12", nil
	}
	return "", fmt.Errorf("unknown CSI ~ escape: %+v", args)
}

func csih(args []int) (string, error) {
	debugf("CSI h")
	switch args[0] {
	case 4: // st sends this
		return "Insert", nil
	}
	return "", fmt.Errorf("unknown CSI h escape: %+v", args)
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
