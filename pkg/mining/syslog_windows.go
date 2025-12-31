//go:build windows

package mining

import (
	"github.com/Snider/Mining/pkg/logging"
)

// On Windows, syslog is not available. We'll use a dummy implementation
// that logs to the standard logger.

// logToSyslog logs a message to the standard logger, mimicking the syslog function's signature.
func logToSyslog(message string) {
	logging.Info(message)
}
