package getkey // import "github.com/mndrix/getkey"
import "fmt"

var isDebug bool

// SetDebug enables or disable debugging log messages.
func SetDebug(enable bool) {
	isDebug = enable
}

func debugf(format string, args ...interface{}) {
	if isDebug {
		fmt.Printf("DEBUG: "+format+"\n", args...)
	}
}
