package installer

import (
	"testing"
)

func TestCheckReleasePath(t *testing.T) {
	tests := []struct {
		name        string
		kernelLine  string
		wantStatus  string
		wantErr     error
		description string
	}{
		{
			name:        "pre-release snapshot (May 4 < release date)",
			kernelLine:  "OpenBSD 7.9-current (GENERIC.MP) #475: Mon May 4 12:34:48 MDT 2026",
			wantStatus:  "pre-release",
			wantErr:     ErrUpgradeRequired,
			description: "old kernel before release, should offer sysupgrade -R",
		},
		{
			name:        "post-release snapshot (May 21 > release date)",
			kernelLine:  "OpenBSD 7.9-current (GENERIC.MP) #480: Thu May 21 10:00:00 MDT 2026",
			wantStatus:  "post-release",
			wantErr:     ErrDowngradeRisk,
			description: "new kernel after release, should warn about downgrade",
		},
		{
			name:        "same-day snapshot (build date == release date)",
			kernelLine:  "OpenBSD 7.9-current (GENERIC.MP) #475: Wed May 6 08:00:00 MDT 2026",
			wantStatus:  "post-release",
			wantErr:     ErrDowngradeRisk,
			description: "same day as release, treated as post-release",
		},
		{
			name:        "stable release (no -current)",
			kernelLine:  "OpenBSD 7.9 (GENERIC.MP) #0: Wed May 6 08:00:00 MDT 2026",
			wantStatus:  "stable",
			wantErr:     nil,
			description: "release build, no migration needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := checkReleasePathFromKernelLine(tt.kernelLine, ReleaseDate)

			if res.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q (%s)", res.Status, tt.wantStatus, tt.description)
			}
			if err != tt.wantErr {
				t.Errorf("err = %v, want %v (%s)", err, tt.wantErr, tt.description)
			}
		})
	}
}

