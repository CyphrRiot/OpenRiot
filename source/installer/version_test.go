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

func TestCheckErrataCurrent(t *testing.T) {
	tests := []struct {
		name          string
		kernelLine    string
		wantUnapplied int
		wantReboot    bool
	}{
		{
			name:          "kernel before errata (May 25 < June 2)",
			kernelLine:    "OpenBSD 7.9-current (GENERIC.MP) #490: Mon May 25 10:00:00 MDT 2026",
			wantUnapplied: 3,
			wantReboot:    true,
		},
		{
			name:          "kernel after errata (June 5 > June 2)",
			kernelLine:    "OpenBSD 7.9-current (GENERIC.MP) #495: Fri Jun 5 10:00:00 MDT 2026",
			wantUnapplied: 0,
			wantReboot:    false,
		},
		{
			name:          "kernel same day as errata (June 2)",
			kernelLine:    "OpenBSD 7.9-current (GENERIC.MP) #492: Tue Jun 2 08:00:00 MDT 2026",
			wantUnapplied: 0,
			wantReboot:    false,
		},
		{
			name:          "bad date format",
			kernelLine:    "OpenBSD 7.9-current (GENERIC.MP) #0",
			wantUnapplied: 0,
			wantReboot:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := checkErrataCurrent(tt.kernelLine)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(res.Unapplied) != tt.wantUnapplied {
				t.Errorf("unapplied = %d, want %d", len(res.Unapplied), tt.wantUnapplied)
			}
			if res.Reboot != tt.wantReboot {
				t.Errorf("reboot = %v, want %v", res.Reboot, tt.wantReboot)
			}
		})
	}
}

func TestParseErrataHTML(t *testing.T) {
	html := `<html><body><ul>
<li id="p001_xserver">
<strong>001: SECURITY FIX: June 2, 2026</strong>
&nbsp; <i>All architectures</i>
<br/>
Multiple vulnerabilites in the X server dri2, sync, saver and Xkb
extensions.
<br/>
<a href="patch.sig">A source code patch exists</a>
</li><li id="p002_smtpd">
<strong>002: RELIABILITY FIX: June 2, 2026</strong>
&nbsp; <i>All architectures</i>
<br/>
Fixes for a variety of crashing bugs in smtpd(8).
<br/>
<a href="patch.sig">A source code patch exists</a>
</li></ul></body></html>`

	entries, err := ParseErrataHTML(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	if entries[0].ID != "001" {
		t.Errorf("entry[0].ID = %q, want %q", entries[0].ID, "001")
	}
	if !entries[0].Reboot {
		t.Errorf("entry[0].Reboot = false, want true (X server)")
	}

	if entries[1].ID != "002" {
		t.Errorf("entry[1].ID = %q, want %q", entries[1].ID, "002")
	}
	if entries[1].Reboot {
		t.Errorf("entry[1].Reboot = true, want false (smtpd)")
	}
}