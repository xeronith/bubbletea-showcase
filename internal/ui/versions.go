package ui

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// DepVersion returns the version of a module baked into the binary, falling
// back to a known value when the build did not record dependency metadata.
func DepVersion(path, fallback string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fallback
	}
	for _, dep := range info.Deps {
		if dep.Path != path {
			continue
		}
		version := dep.Version
		if dep.Replace != nil && dep.Replace.Version != "" {
			version = dep.Replace.Version
		}
		if version != "" && version != "(devel)" {
			return strings.TrimPrefix(version, "v")
		}
	}
	return fallback
}

// GoVersion returns the toolchain version without the "go" prefix.
func GoVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}
