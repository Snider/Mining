//go:build windows

package mining

import (
	"log"
)

// On Windows, syslog is not available. We'll use a dummy implementation
// that logs to the standard logger.

// logToSyslog logs a message to the standard logger, mimicking the syslog function's signature.
func logToSyslog(message string) {
	log.Println(message)
}
