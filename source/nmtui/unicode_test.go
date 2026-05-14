package nmtui

import "testing"

func TestUnicodeSSIDs(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		// Unquoted Unicode SSIDs
		{`nwid Gödel chan 6 bssid de:ad:be:ef:ca:fe 83% HT-MCS0 privacy,wpa2`, "Gödel"},
		{`nwid Café chan 1 bssid 00:11:22:33:44:55 65% HT-MCS0 wpa2`, "Café"},
		{`nwid Hütte chan 11 bssid aa:bb:cc:dd:ee:ff 42% open`, "Hütte"},
		{`nwid Röd chan 1 bssid 11:22:33:44:55:66 50% wpa2`, "Röd"},
		// Quoted Unicode SSIDs with spaces
		{`nwid "Gödel's Network" chan 6 bssid ab:cd:ef:01:23:45 77% wpa2`, "Gödel's Network"},
		{`nwid "Café Noir" chan 1 bssid bb:cc:dd:ee:ff:00 60% wpa2`, "Café Noir"},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			aps := parseScanOutput(tc.line)
			if len(aps) == 0 {
				t.Fatalf("got 0 APs, want 1")
			}
			if aps[0].SSID != tc.want {
				t.Errorf("SSID = %q, want %q", aps[0].SSID, tc.want)
			}
		})
	}
}
