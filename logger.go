package munn

import "fmt"

// DEBUG if true will enable debug logging
var DEBUG = false

func logDebug(format string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(format, args...)
	}
}
