// Package buildinfo exposes immutable build metadata injected through linker flags.
package buildinfo

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// Info is safe to expose through the public metadata endpoint.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// Current returns the metadata compiled into this binary.
func Current() Info {
	return Info{Version: version, Commit: commit, BuildTime: buildTime}
}
