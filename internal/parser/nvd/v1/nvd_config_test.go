package v1

import (
	"testing"
)

func TestParseCpe23Uri(t *testing.T) {
	tests := []struct {
		in          string
		wantVendor  string
		wantProduct string
	}{
		{"cpe:2.3:a:oracle:mysql:5.7.18:*:*:*:*:*:*:*", "oracle", "mysql"},
		{"cpe:2.3:a:microsoft:office:2016:*:*:*:*:*:*:*", "microsoft", "office"},
		{"invalid", "", ""},
		{"cpe:2.3:a:only", "", ""},
	}
	for _, tt := range tests {
		vendor, product := parseCpe23Uri(tt.in)
		if vendor != tt.wantVendor || product != tt.wantProduct {
			t.Errorf("parseCpe23Uri(%q) = %q,%q want %q,%q", tt.in, vendor, product, tt.wantVendor, tt.wantProduct)
		}
	}
}

func TestFlattenNodes(t *testing.T) {
	grandchild := ConfigNode{CpeMatch: []CpeMatch{{Cpe23Uri: "grandchild"}}}
	child1 := ConfigNode{CpeMatch: []CpeMatch{{Cpe23Uri: "child1"}}}
	child2 := ConfigNode{CpeMatch: []CpeMatch{{Cpe23Uri: "child2"}}, Children: []ConfigNode{grandchild}}
	root1 := ConfigNode{CpeMatch: []CpeMatch{{Cpe23Uri: "root1"}}, Children: []ConfigNode{child1, child2}}
	root2 := ConfigNode{CpeMatch: []CpeMatch{{Cpe23Uri: "root2"}}}

	nodes := []ConfigNode{root1, root2}
	flat := flattenNodes(nodes)

	wantOrder := []string{"root1", "child1", "child2", "grandchild", "root2"}
	if len(flat) != len(wantOrder) {
		t.Fatalf("flattenNodes returned %d nodes, want %d", len(flat), len(wantOrder))
	}
	for i, n := range flat {
		got := ""
		if len(n.CpeMatch) > 0 {
			got = n.CpeMatch[0].Cpe23Uri
		}
		if got != wantOrder[i] {
			t.Errorf("node %d uri = %q, want %q", i, got, wantOrder[i])
		}
	}
}
