package nmtui

import (
	"fmt"
	"testing"
)

func TestFindWiFiInterface(t *testing.T) {
	iface, err := FindWiFiInterface()
	if err != nil {
		t.Logf("no wifi interface found (expected on systems without wifi): %v", err)
		return
	}
	if iface == "" {
		t.Fatal("expected non-empty interface name")
	}
	t.Logf("found interface: %s", iface)
}

func TestParseScanOutput(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantLen int
		firstAP WiFiAP
	}{
		{
			name: "single_wpa2",
			input: `
				nwid MyHome chan 6 bssid 12:34:56:78:9a:bc 83% HT-MCS0 privacy,short_preamble,wpa2
			`,
			wantLen: 1,
			firstAP: WiFiAP{
				SSID:        "MyHome",
				BSSID:       "12:34:56:78:9a:bc",
				Signal:      83,
				SignalValid: true,
				Security:    "wpa2",
			},
		},
		{
			name: "open_network",
			input: `
				nwid CoffeeShop chan 1 bssid aa:bb:cc:dd:ee:ff 65% HT-MCS0 radio_measurement
			`,
			wantLen: 1,
			firstAP: WiFiAP{
				SSID:        "CoffeeShop",
				BSSID:       "aa:bb:cc:dd:ee:ff",
				Signal:      65,
				SignalValid: true,
				Security:    "open",
			},
		},
		{
			name: "quoted_ssid_with_spaces",
			input: `
				nwid "Cox Mobile" chan 157 bssid ab:cd:ef:01:23:45 73% HT-MCS71 privacy,spectrum_mgmt,wpa2,802.1x
			`,
			wantLen: 1,
			firstAP: WiFiAP{
				SSID:        "Cox Mobile",
				BSSID:       "ab:cd:ef:01:23:45",
				Signal:      73,
				SignalValid: true,
				Security:    "wpa2",
			},
		},
		{
			name: "wpa3",
			input: `
				nwid NETGEAR chan 1 bssid de:ad:be:ef:ca:fe 47% HT-MCS31 privacy,short_slottime,wpa3,wpa2
			`,
			wantLen: 1,
			firstAP: WiFiAP{
				SSID:        "NETGEAR",
				BSSID:       "de:ad:be:ef:ca:fe",
				Signal:      47,
				SignalValid: true,
				Security:    "wpa3",
			},
		},
		{
			name: "wep",
			input: `
				nwid OldNet chan 1 bssid 00:11:22:33:44:55 64% HT-MCS23 privacy,radio_measurement,wep
			`,
			wantLen: 1,
			firstAP: WiFiAP{
				SSID:        "OldNet",
				BSSID:       "00:11:22:33:44:55",
				Signal:      64,
				SignalValid: true,
				Security:    "wep",
			},
		},
		{
			name:    "empty_output",
			input:   "",
			wantLen: 0,
		},
		{
			name: "hidden_filtered_out",
			input: `
				nwid "" chan 44 bssid 8e:76:3f:7b:df:92 82% HT-MCS31 privacy,radio_measurement,wpa2
			`,
			wantLen: 0, // hidden networks skipped
		},
		{
			name: "realistic_mixed",
			input: `
				iwx0: flags=808843<UP,BROADCAST,RUNNING> mtu 1500
					lladdr bc:76:37:d3:61:2d
					media: IEEE802.11 autoselect (VHT-MCS0 mode 11ac)
					status: active
					nwid SETUP-0B0A chan 157 bssid 88:9e:68:9c:0b:16 83% HT-MCS71 privacy,wpa2
					nwid SETUP-0B0A chan 11 bssid 88:9e:68:9c:0b:0e 82% HT-MCS39 privacy,wpa2
					nwid "Cox Mobile" chan 157 bssid 88:9e:68:9c:0b:1a 73% HT-MCS71 privacy,wpa2,802.1x
					nwid CoxWiFi chan 44 bssid 8a:76:3f:7b:df:92 80% HT-MCS31 radio_measurement
					nwid "" chan 44 bssid 8e:76:3f:7b:df:92 82% HT-MCS31 privacy,wpa2
					inet 192.168.0.44 netmask 0xffffff00
			`,
			wantLen: 3,
			firstAP: WiFiAP{ // first unique nwid line
				SSID:        "SETUP-0B0A",
				BSSID:       "88:9e:68:9c:0b:16",
				Signal:      83,
				SignalValid: true,
				Security:    "wpa2",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseScanOutput(tc.input)
			if len(got) != tc.wantLen {
				t.Fatalf("got %d APs, want %d", len(got), tc.wantLen)
			}
			if tc.wantLen == 0 {
				return
			}
			ap := got[0]
			if ap.SSID != tc.firstAP.SSID {
				t.Errorf("SSID = %q, want %q", ap.SSID, tc.firstAP.SSID)
			}
			if ap.BSSID != tc.firstAP.BSSID {
				t.Errorf("BSSID = %q, want %q", ap.BSSID, tc.firstAP.BSSID)
			}
			if ap.Signal != tc.firstAP.Signal {
				t.Errorf("Signal = %d, want %d", ap.Signal, tc.firstAP.Signal)
			}
			if ap.Security != tc.firstAP.Security {
				t.Errorf("Security = %q, want %q", ap.Security, tc.firstAP.Security)
			}
		})
	}
}

func TestParseScanOutputDeduplicate(t *testing.T) {
	input := `
		nwid SETUP-0B0A chan 157 bssid 88:9e:68:9c:0b:16 83% HT-MCS71 privacy,wpa2
		nwid SETUP-0B0A chan 11 bssid 88:9e:68:9c:0b:0e 82% HT-MCS39 privacy,wpa2
		nwid SETUP-0B0A chan 44 bssid 88:9e:68:9c:0b:0f 77% HT-MCS0 privacy,wpa2
	`
	got := parseScanOutput(input)
	if len(got) != 1 {
		t.Fatalf("expected 1 deduplicated AP, got %d", len(got))
	}
	if got[0].SSID != "SETUP-0B0A" {
		t.Errorf("SSID = %q, want SETUP-0B0A", got[0].SSID)
	}
	if got[0].Signal != 83 {
		t.Errorf("Signal = %d, want 83 (strongest)", got[0].Signal)
	}
	if got[0].BSSID != "88:9e:68:9c:0b:16" {
		t.Errorf("BSSID = %q, want 88:9e:68:9c:0b:16", got[0].BSSID)
	}
}

func TestExtractSSID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`MyHome chan 6 bssid ...`, "MyHome"},
		{`"Cox Mobile" chan 157 ...`, "Cox Mobile"},
		{`"" chan 3 ...`, ""},
		{`Guest open`, "Guest"},
		{`bssid 00:11:22:33:44:55 ...`, ""},
		{``, ""},
		{`83% HT-MCS0 privacy,wpa2`, ""},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := extractSSID(tc.input)
			if got != tc.want {
				t.Errorf("extractSSID(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSignalToPercent(t *testing.T) {
	cases := []struct {
		signal int
		valid  bool
		want   int
	}{
		{83, true, 83},
		{100, true, 100},
		{0, true, 0},
		{42, true, 42},
		{105, true, 100},
		{-5, true, 0},
		{0, false, -1},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d_valid=%v", tc.signal, tc.valid), func(t *testing.T) {
			got := SignalToPercent(tc.signal, tc.valid)
			if got != tc.want {
				t.Errorf("SignalToPercent(%d, %v) = %d, want %d", tc.signal, tc.valid, got, tc.want)
			}
		})
	}
}

func TestParseIfconfigConnection(t *testing.T) {
	input := `
iwx0: flags=808843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST,AUTOCONF4> mtu 1500
	lladdr bc:76:37:d3:61:2d
	index 1 priority 4 llprio 3
	groups: wlan egress
	media: IEEE802.11 autoselect (VHT-MCS0 mode 11ac)
	status: active
	ieee80211: join SETUP-0B0A chan 157 bssid 88:9e:68:9c:0b:16 77% wpakey wpaprotos wpa2 wpaakms psk,sha256-psk wpaciphers ccmp wpagroupcipher ccmp powersave on
		ssid SETUP-0B0A channel 5785 (5745/160) bssid 88:9e:68:9c:0b:16
		country US ecm
		vaesad
		nullfunc
		powersavesleep 100
		powersave
		txpower 22
		bintval 100
		rateset default
		vhtparams: rxldpc 2x2:2
		htconf ht40+
		vhtconf vht160
		ampdu
		amsdu
		protmodects
		ibss
		nwflag WPS
		pureg
		puren
		wme
		wep
		txop
		roaming device
		bssid 88:9e:68:9c:0b:16
		null
		11n
		11ac
		11g
		11b
		11a
	inet 192.168.0.44 netmask 0xffffff00 broadcast 192.168.0.255
	inet6 fe80::be76:37ff:fed3:612d%iwx0 prefixlen 64 scopeid 0x1
`
	got := parseIfconfigConnection(input)
	if got.SSID != "SETUP-0B0A" {
		t.Errorf("SSID = %q, want SETUP-0B0A", got.SSID)
	}
	if got.IP != "192.168.0.44" {
		t.Errorf("IP = %q, want 192.168.0.44", got.IP)
	}
	if got.MAC != "bc:76:37:d3:61:2d" {
		t.Errorf("MAC = %q, want bc:76:37:d3:61:2d", got.MAC)
	}
	if got.State != "active" {
		t.Errorf("State = %q, want active", got.State)
	}
}

func TestGetKnownSSIDs(t *testing.T) {
	iface, err := FindWiFiInterface()
	if err != nil {
		t.Skip("no wifi interface found")
	}
	ssids := GetKnownSSIDs(iface)
	t.Logf("known SSIDs for %s: %v", iface, ssids)
}

func TestIsWiFiConnected(t *testing.T) {
	iface, err := FindWiFiInterface()
	if err != nil {
		t.Skip("no wifi interface found")
	}
	connected := IsWiFiConnected(iface)
	t.Logf("interface %s connected: %v", iface, connected)
}
