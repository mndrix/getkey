package getkey // import "github.com/mndrix/getkey"
import (
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/pkg/term"
)

var terminal *term.Term
var envTerm string
var buf []byte
var mux sync.Mutex

func init() {
	envTerm = os.Getenv("TERM")
	buf = make([]byte, 15)
}

func getTerminal() *term.Term {
	if terminal == nil {
		var err error
		terminal, err = term.Open("/dev/tty")
		if err != nil {
			panic(err)
		}
	}
	return terminal
}

// should only be called while holding mutex
func prepare() {
	// TODO should probably use termcap or something
	if envTerm == "xterm" || strings.HasPrefix(envTerm, "xterm-") {
		//fmt.Fprint(terminal, "\x1b[>4;2m") // xterm: set modifyOtherKeys=2
		//fmt.Print("x")                     // xterm eats first character
	}
}

// should only be called while holding mutex
func restore() {
	if envTerm == "xterm" || strings.HasPrefix(envTerm, "xterm-") {
		//fmt.Fprint(terminal, "\x1b[>4m") // xterm: restore modifyOtherKeys
	}
}

// returns a sequence of raw bytes read from the terminal.
func read() ([]byte, error) {
	mux.Lock()
	defer mux.Unlock()

	term.RawMode(getTerminal())
	defer getTerminal().Restore()

	prepare()
	defer restore()

	numRead, err := getTerminal().Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "reading")
	}

	return buf[0:numRead], nil
}

// GetKey waits for the user to press a key and then returns a string
// describing what was pressed.  Alphanumeric characters and
// punctuation are represented as themselves.  Modifier keys are
// represented as Ctrl-, Alt-, Esc-, and Shift- prefixes (in that
// order) on the base key.  For example, holding down Control and Alt
// while pressing the L key produces "Ctrl-Alt-L".
func GetKey() (string, error) {
	raw, err := read()
	if err != nil {
		return "", errors.Wrap(err, "GetKey")
	}
	return decode(raw)
}

// Restore restores the terminal to the state it was in during init().
//
// You don't usually need to call this function because GetKey calls
// it automatically when it returns.  However, it can be helpful to
// call during program exit if GetKey was running in a separate
// goroutine.
func Restore() (err error) {
	err = getTerminal().Restore()
	restore()
	return
}
