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

// GetCommit returns the git commit hash
func GetCommit() string {
	return commit
}

// GetBuildDate returns the build date
func GetBuildDate() string {
	return date
}
