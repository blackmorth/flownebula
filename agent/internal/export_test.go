package internal

import "testing"

func TestParseNodeNameVariants(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		nodeID     string
		display    string
		calledClas string
	}{
		{
			name:       "closure",
			raw:        "{closure}::index.php/12",
			nodeID:     "{closure}::index.php/12",
			display:    "",
			calledClas: "",
		},
		{
			name:       "class method",
			raw:        "index.php::Service::Run",
			nodeID:     "Service::Run",
			display:    "Service::Run",
			calledClas: "Service",
		},
		{
			name:       "script level unknown function",
			raw:        "path/to/index.php::Main::unknown",
			nodeID:     "run_init::index.php",
			display:    "",
			calledClas: "",
		},
		{
			name:       "global function",
			raw:        "index.php::::strtolower",
			nodeID:     "strtolower",
			display:    "",
			calledClas: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeID, display, cls := parseNodeName(tt.raw)
			if nodeID != tt.nodeID || display != tt.display || cls != tt.calledClas {
				t.Fatalf("parseNodeName(%q) = (%q,%q,%q), expected (%q,%q,%q)", tt.raw, nodeID, display, cls, tt.nodeID, tt.display, tt.calledClas)
			}
		})
	}
}

func TestRound3(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0, 0},
		{12.3456, 12.3},
		{9999.9, 10000},
		{0.012345, 0.0123},
	}

	for _, tc := range cases {
		got := round3(tc.in)
		if got != tc.want {
			t.Fatalf("round3(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
