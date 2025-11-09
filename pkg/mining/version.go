package mining

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// GetVersion returns the version of the application
func GetVersion() string {
	return version
}
