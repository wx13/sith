package version

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
)

// pseudoVersionRe matches Go pseudo-versions like "v1.0.1-0.20260409215840-6eb57b80b456"
var pseudoVersionRe = regexp.MustCompile(`^(v\d+\.\d+\.)(\d+)-0\.\d{14}-[0-9a-f]{12}`)

// Get returns a formatted version string.
func Get() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}

	var revision string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}

	version := info.Main.Version

	// Parse pseudo-versions to extract base tag
	if m := pseudoVersionRe.FindStringSubmatch(version); m != nil {
		// m[1] = "v1.0.", m[2] = "1" (incremented patch)
		// Decrement patch to get actual base tag
		patch, _ := strconv.Atoi(m[2])
		if patch > 0 {
			version = fmt.Sprintf("%s%d", m[1], patch-1)
		}
	}

	if version == "" || version == "(devel)" {
		version = "dev"
	}

	// Add short commit hash
	if len(revision) >= 7 {
		version = fmt.Sprintf("%s (%s)", version, revision[:7])
	}

	if dirty {
		version += " dirty"
	}

	return version
}
