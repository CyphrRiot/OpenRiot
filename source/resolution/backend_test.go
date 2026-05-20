package resolution

import (
	"testing"
)

func TestParseXrandr(t *testing.T) {
	sample := `eDP-1 connected primary (normal left inverted right axis y axis)
   1920x1080     60.00 +  60.00* 
   1680x1050     60.00  
HDMI-1 connected 1920x1080+0+0
   1920x1080     75.00*+  60.00    50.00  
   1600x900      60.00  
DP-1 disconnected`

	displays := parseXrandr(sample)

	if len(displays) != 2 {
		t.Fatalf("expected 2 displays, got %d", len(displays))
	}

	eDP := displays[0]
	if eDP.Name != "eDP-1" {
		t.Errorf("expected eDP-1, got %s", eDP.Name)
	}
	if !eDP.Primary {
		t.Error("expected eDP-1 to be primary")
	}

	// Check current resolution
	if eDP.Current != "1920x1080@60.00" {
		t.Errorf("expected current 1920x1080@60.00, got %s", eDP.Current)
	}

	// Check first mode
	if len(eDP.Modes) < 2 {
		t.Fatalf("expected at least 2 modes for eDP-1, got %d", len(eDP.Modes))
	}
	m0 := eDP.Modes[0]
	if m0.Resolution != "1920x1080" {
		t.Errorf("expected resolution 1920x1080, got %s", m0.Resolution)
	}
	if len(m0.Rates) != 2 {
		t.Fatalf("expected 2 rates, got %d", len(m0.Rates))
	}
	if !m0.Rates[1].Current {
		t.Error("expected second rate (60.00*) to be current")
	}
	if !m0.Rates[0].Preferred {
		t.Error("expected first rate (60.00+) to be preferred")
	}

	// Check second display
	hdmi := displays[1]
	if hdmi.Name != "HDMI-1" {
		t.Errorf("expected HDMI-1, got %s", hdmi.Name)
	}
	if hdmi.Primary {
		t.Error("expected HDMI-1 to not be primary")
	}

	// Check HDMI rates
	if len(hdmi.Modes) == 0 {
		t.Fatal("expected HDMI-1 to have modes")
	}
	r := hdmi.Modes[0].Rates[0]
	if !r.Current && !r.Preferred {
		t.Error("expected first HDMI rate to be both current and preferred")
	}
}

func TestRateString(t *testing.T) {
	tests := []struct {
		rate Rate
		want string
	}{
		{Rate{Value: 60.0}, "60.00"},
		{Rate{Value: 60.0, Current: true}, "60.00 *"},
		{Rate{Value: 75.0, Preferred: true}, "75.00 +"},
		{Rate{Value: 60.0, Current: true, Preferred: true}, "60.00 * +"},
	}
	for _, tt := range tests {
		got := tt.rate.String()
		if got != tt.want {
			t.Errorf("Rate(%+v).String() = %q, want %q", tt.rate, got, tt.want)
		}
	}
}
