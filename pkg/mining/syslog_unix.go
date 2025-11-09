//go:build !windows

package mining

import (
	"log"
	"log/syslog"
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
		log.Printf("Failed to connect to syslog: %v. Syslog logging will be disabled.", err)
		syslogWriter = nil // Ensure it's nil on failure
	}
}

// logToSyslog sends a message to syslog if available, otherwise falls back to standard log.
func logToSyslog(message string) {
	if syslogWriter != nil {
		_ = syslogWriter.Notice(message)
	} else {
		log.Println(message)
	}
}
