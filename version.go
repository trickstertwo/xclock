package xclock

import "runtime/debug"

// Version returns the module version as recorded in the build info.
// When built from a tagged module as a dependency, it returns that tag (e.g., v0.2.0).
// When built locally (replace, no tag), it returns "devel".
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	// If xclock is the main module (rare for libraries)
	if info.Main.Path == "github.com/trickstertwo/xclock" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	// When xclock is a dependency (normal for consumers)
	for _, dep := range info.Deps {
		if dep.Path == "github.com/trickstertwo/xclock" && dep.Version != "" {
			return dep.Version
		}
	}
	return "devel"
}
