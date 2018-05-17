package getkey // import "github.com/mndrix/getkey"
import (
	"fmt"
	"os"
)

var isDebug bool

// SetDebug enables or disable debugging log messages.
func SetDebug(enable bool) {
	isDebug = enable
}

func debugf(format string, args ...interface{}) {
	if isDebug {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
	}
}
