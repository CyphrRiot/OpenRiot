//go:build openbsd

package migrate

// DefaultMountPoint returns the default mount point prefix on OpenBSD.
func DefaultMountPoint() string {
	return "/mnt"
}
