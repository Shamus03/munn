package munn

import "fmt"

func (p *Portfolio) logDebug(format string, args ...interface{}) {
	if p.Debug {
		fmt.Printf(format, args...)
	}
}
