package jpm

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		major int
		minor int
		patch int
		pre   string
		err   bool
	}{
		{"1.2.3", 1, 2, 3, "", false},
		{"v2.0.1", 2, 0, 1, "", false},
		{"0.0.0", 0, 0, 0, "", false},
		{"1.2.3-beta", 1, 2, 3, "beta", false},
		{"1.2.3-beta.1", 1, 2, 3, "beta.1", false},
		{"1.2.3+build.42", 1, 2, 3, "", false},
		{"1.2.3-beta+build.42", 1, 2, 3, "beta", false},
		{"abc", 0, 0, 0, "", true},
	}
	for _, tt := range tests {
		v, err := ParseVersion(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("ParseVersion(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseVersion(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
			t.Errorf("ParseVersion(%q) = %d.%d.%d, want %d.%d.%d",
				tt.input, v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
		}
		if v.PreRelease != tt.pre {
			t.Errorf("ParseVersion(%q) pre = %q, want %q", tt.input, v.PreRelease, tt.pre)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.2.0", "1.1.0", 1},
		{"1.1.1", "1.1.0", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
	}
	for _, tt := range tests {
		a := MustParseVersion(tt.a)
		b := MustParseVersion(tt.b)
		got := a.Compare(b)
		if got != tt.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestConstraintCaret(t *testing.T) {
	c, err := ParseConstraint("^1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Matches(MustParseVersion("1.2.3")) {
		t.Error("^1.2.3 should match 1.2.3")
	}
	if !c.Matches(MustParseVersion("1.9.9")) {
		t.Error("^1.2.3 should match 1.9.9")
	}
	if c.Matches(MustParseVersion("2.0.0")) {
		t.Error("^1.2.3 should NOT match 2.0.0")
	}
	if c.Matches(MustParseVersion("1.2.2")) {
		t.Error("^1.2.3 should NOT match 1.2.2")
	}
}

func TestConstraintTilde(t *testing.T) {
	c, err := ParseConstraint("~1.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Matches(MustParseVersion("1.2.0")) {
		t.Error("~1.2.0 should match 1.2.0")
	}
	if !c.Matches(MustParseVersion("1.2.9")) {
		t.Error("~1.2.0 should match 1.2.9")
	}
	if c.Matches(MustParseVersion("1.3.0")) {
		t.Error("~1.2.0 should NOT match 1.3.0")
	}
}

func TestConstraintExact(t *testing.T) {
	c, err := ParseConstraint("1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Matches(MustParseVersion("1.2.3")) {
		t.Error("exact should match 1.2.3")
	}
	if c.Matches(MustParseVersion("1.2.4")) {
		t.Error("exact should NOT match 1.2.4")
	}
}

func TestConstraintRange(t *testing.T) {
	c, err := ParseConstraint(">=1.0.0 <2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Matches(MustParseVersion("1.5.0")) {
		t.Error("range should match 1.5.0")
	}
	if c.Matches(MustParseVersion("2.0.0")) {
		t.Error("range should NOT match 2.0.0")
	}
	if c.Matches(MustParseVersion("0.9.0")) {
		t.Error("range should NOT match 0.9.0")
	}
}

func TestBestMatch(t *testing.T) {
	c := &Constraint{ops: []constraintOp{{operator: "^", version: MustParseVersion("1.0.0")}}}
	versions := VersionsFromStrings([]string{"0.9.0", "1.0.0", "1.2.0", "1.5.0", "2.0.0"})
	v, ok := BestMatch(c, versions)
	if !ok {
		t.Fatal("expected best match")
	}
	if v.String() != "1.5.0" {
		t.Errorf("BestMatch = %s, want 1.5.0", v.String())
	}
}

func TestBestMatch_NoMatch(t *testing.T) {
	c := &Constraint{ops: []constraintOp{{operator: "^", version: MustParseVersion("3.0.0")}}}
	versions := VersionsFromStrings([]string{"1.0.0", "2.0.0"})
	_, ok := BestMatch(c, versions)
	if ok {
		t.Error("expected no match")
	}
}

func TestCoerce(t *testing.T) {
	v, err := Coerce("v1.2")
	if err != nil {
		t.Fatal(err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 0 {
		t.Errorf("Coerce(v1.2) = %s, want 1.2.0", v.String())
	}
}

func TestVersionString(t *testing.T) {
	v := MustParseVersion("1.2.3-beta.1+build.42")
	want := "1.2.3-beta.1+build.42"
	if v.String() != want {
		t.Errorf("Version.String() = %q, want %q", v.String(), want)
	}
}
