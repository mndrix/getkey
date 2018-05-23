package getkey // import "github.com/mndrix/getkey"
import (
	"fmt"
)

// map terminal escape sequnces to their descriptive names
var m map[string]string

func init() {
	m = make(map[string]string, 256*4)

	// initial control sequences can be generated with Ctrl
	for i := 0; i <= 31; i++ {
		k := string([]rune{rune(i)})
		offset := 96
		if i >= 27 {
			offset = 64
		}
		v := fmt.Sprintf("Ctrl-%c", i+offset)
		//fmt.Fprintf(os.Stderr, "mapping %q to %q\n", k, v)
		m[k] = v

		// same with Alt-Ctrl
		m["\x1b"+k] = fmt.Sprintf("Ctrl-Alt-%c", i+offset)
	}

	// printable ASCII is represented as itself
	for i := 33; i <= 126; i++ {
		s := fmt.Sprintf("%c", i)
		m[s] = s
		m["\x1b"+s] = "Alt-" + s
	}

	// CSI
	csi := "\x1b["
	m[csi+"A"] = "Up"
	m[csi+"B"] = "Down"
	m[csi+"C"] = "Right"
	m[csi+"D"] = "Left"
	m[csi+"F"] = "End"       // xterm
	m[csi+"H"] = "Home"      // xterm
	m[csi+"P"] = "Backspace" // st

	// CSI h
	m[csi+"4h"] = "Insert"

	// CSI ~
	m[csi+"1~"] = "Home"
	m[csi+"2~"] = "Insert"
	m[csi+"3~"] = "Delete"
	m[csi+"4~"] = "End"
	m[csi+"5~"] = "PageUp"
	m[csi+"6~"] = "PageDown"
	m[csi+"15~"] = "F5"
	m[csi+"17~"] = "F6"
	m[csi+"18~"] = "F7"
	m[csi+"19~"] = "F8"
	m[csi+"20~"] = "F9"
	m[csi+"21~"] = "F10"
	m[csi+"23~"] = "F11"
	m[csi+"24~"] = "F12"

	// some function keys
	z := "\x1bO"
	m[z+"P"] = "F1"
	m[z+"Q"] = "F2"
	m[z+"R"] = "F3"
	m[z+"S"] = "F4"

	// handle some manually (including overwriting mistakes above)
	m["\x1b"] = "Escape"
	m[" "] = "Space"
	m["\r"] = "Enter"
	m["\t"] = "Tab"
	m["\x00"] = "Ctrl-Space"
	m["\x15"] = "Ctrl-u"
}

// decode a single sequence of raw bytes from the terminal
func decode(c []byte) (string, error) {
	s := string(c)
	debugf("raw: %v %q", c, s)
	if name, ok := m[s]; ok {
		debugf("shortcut")
		return name, nil
	}
	return "(unknown)", nil
}
