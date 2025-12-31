//go:build !windows

package mining

import (
	"log/syslog"

	"github.com/Snider/Mining/pkg/logging"
)

var syslogWriter *syslog.Writer

func init() {
	// Initialize syslog writer globally.
	// LOG_NOTICE is for normal but significant condition.
	// LOG_DAEMON is for system daemons.
	// "mining-service" is the tag for the log messages.
	var err error
	syslogWriter, err = syslog.New(syslog.LOG_NOTICE|syslog.LOG_DAEMON, "mining-service")
	if err != nil {
		logging.Warn("failed to connect to syslog, syslog logging disabled", logging.Fields{"error": err})
		syslogWriter = nil // Ensure it's nil on failure
	}
}

// logToSyslog sends a message to syslog if available, otherwise falls back to standard log.
func logToSyslog(message string) {
	if syslogWriter != nil {
		_ = syslogWriter.Notice(message)
	} else {
		logging.Info(message)
	}
}
