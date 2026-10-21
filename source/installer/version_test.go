package installer

import (
	"testing"
	"time"
)

func TestCompareRelease(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"7.9", "8.0", -1},
		{"8.0", "8.0", 0},
		{"8.1", "8.0", 1},
		{"7.10", "7.9", 1},
	}
	for _, tt := range tests {
		if got := CompareRelease(tt.a, tt.b); got != tt.want {
			t.Errorf("CompareRelease(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCheckReleasePath(t *testing.T) {
	releaseDate := time.Date(2026, time.October, 6, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		kernelLine  string
		wantStatus  string
		wantErr     error
		description string
	}{
		{
			name:        "pre-release snapshot (Oct 5 < release date)",
			kernelLine:  "OpenBSD 8.0-current (GENERIC.MP) #475: Mon Oct 5 12:34:48 MDT 2026",
			wantStatus:  "pre-release",
			wantErr:     ErrUpgradeRequired,
			description: "old kernel before release, should offer sysupgrade -R",
		},
		{
			name:        "post-release snapshot (Oct 8 > release date)",
			kernelLine:  "OpenBSD 8.0-current (GENERIC.MP) #480: Thu Oct 8 10:00:00 MDT 2026",
			wantStatus:  "post-release",
			wantErr:     ErrDowngradeRisk,
			description: "new kernel after release, should warn about downgrade",
		},
		{
			name:        "same-day snapshot (build date == release date)",
			kernelLine:  "OpenBSD 8.0-current (GENERIC.MP) #475: Tue Oct 6 08:00:00 MDT 2026",
			wantStatus:  "post-release",
			wantErr:     ErrDowngradeRisk,
			description: "same day as release, treated as post-release",
		},
		{
			name:        "stable release (no -current)",
			kernelLine:  "OpenBSD 8.0 (GENERIC.MP) #0: Tue Oct 6 08:00:00 MDT 2026",
			wantStatus:  "stable",
			wantErr:     nil,
			description: "release build, no migration needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := checkReleasePathFromKernelLine(tt.kernelLine, releaseDate)

			if res.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q (%s)", res.Status, tt.wantStatus, tt.description)
			}
			if err != tt.wantErr {
				t.Errorf("err = %v, want %v (%s)", err, tt.wantErr, tt.description)
			}
		})
	}
}