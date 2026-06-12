package jpm

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Version represents a parsed semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string // e.g. "alpha", "beta.1", ""
	Build      string // e.g. "build.42", ""
}

// ParseVersion parses a semver string like "1.2.3", "1.2.3-beta", "1.2.3+build.1".
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	re := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z.-]+))?(?:\+([0-9A-Za-z.-]+))?$`)
	parts := re.FindStringSubmatch(s)
	if parts == nil {
		return Version{}, fmt.Errorf("invalid semver: %q", s)
	}
	v := Version{}
	v.Major, _ = strconv.Atoi(parts[1])
	if parts[2] != "" {
		v.Minor, _ = strconv.Atoi(parts[2])
	}
	if parts[3] != "" {
		v.Patch, _ = strconv.Atoi(parts[3])
	}
	v.PreRelease = parts[4]
	v.Build = parts[5]
	return v, nil
}

// String returns the string representation.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare returns -1 if v < other, 0 if equal, 1 if v > other.
// Build metadata is ignored for precedence.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return cmpInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return cmpInt(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return cmpInt(v.Patch, other.Patch)
	}
	// Pre-release: no pre-release > any pre-release
	if v.PreRelease == "" && other.PreRelease != "" {
		return 1
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		return -1
	}
	return comparePreRelease(v.PreRelease, other.PreRelease)
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func comparePreRelease(a, b string) int {
	if a == b {
		return 0
	}
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		c := cmpIdentifier(aParts[i], bParts[i])
		if c != 0 {
			return c
		}
	}
	return cmpInt(len(aParts), len(bParts))
}

func cmpIdentifier(a, b string) int {
	aNum, aErr := strconv.Atoi(a)
	bNum, bErr := strconv.Atoi(b)
	if aErr == nil && bErr == nil {
		return cmpInt(aNum, bNum)
	}
	if aErr == nil {
		return -1 // numeric < non-numeric
	}
	if bErr == nil {
		return 1
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// Constraint represents a semver constraint (e.g. "^1.2.3", "~1.2", ">=1.0 <2.0").
type Constraint struct {
	raw string
	ops []constraintOp
}

type constraintOp struct {
	operator string // "", "=", ">", ">=", "<", "<=", "^", "~"
	version  Version
}

// ParseConstraint parses a semver constraint string.
func ParseConstraint(s string) (*Constraint, error) {
	c := &Constraint{raw: s}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty constraint")
	}

	ops := make([]string, 0)
	verStrs := make([]string, 0)

	for _, part := range parts {
		op, ver := splitOperator(part)
		ops = append(ops, op)
		verStrs = append(verStrs, ver)
	}

	if len(ops) == 1 && ops[0] == "" {
		op := ""
		verStr := verStrs[0]
		if verStr[0] == '^' || verStr[0] == '~' {
			op = string(verStr[0])
			verStr = verStr[1:]
		}
		v, err := ParseVersion(verStr)
		if err != nil {
			return nil, err
		}
		c.ops = append(c.ops, constraintOp{operator: op, version: v})
		return c, nil
	}

	for i := 0; i < len(ops); i++ {
		if ops[i] == "" {
			continue
		}
		v, err := ParseVersion(verStrs[i])
		if err != nil {
			return nil, err
		}
		if ops[i] != ">" && ops[i] != ">=" && ops[i] != "<" && ops[i] != "<=" && ops[i] != "=" {
			return nil, fmt.Errorf("unsupported constraint operator: %q", ops[i])
		}
		c.ops = append(c.ops, constraintOp{operator: ops[i], version: v})
	}

	if len(c.ops) == 0 {
		return nil, fmt.Errorf("malformed constraint: %q", s)
	}
	return c, nil
}

// splitOperator separates ">=1.0.0" into (">=", "1.0.0").
func splitOperator(s string) (string, string) {
	if len(s) >= 2 {
		prefix := s[:2]
		if prefix == ">=" || prefix == "<=" || prefix == "==" {
			return prefix, s[2:]
		}
	}
	if len(s) >= 1 {
		switch s[0] {
		case '>', '<', '=':
			return string(s[0]), s[1:]
		case '^', '~':
			return "", s // ^/~ are part of version, handled later
		}
	}
	return "", s
}

// Matches checks if a version satisfies the constraint.
func (c *Constraint) Matches(v Version) bool {
	for _, op := range c.ops {
		switch op.operator {
		case "":
			// Exact match (pre-release must match too)
			if v.Compare(op.version) != 0 {
				return false
			}
			if v.PreRelease != op.version.PreRelease {
				return false
			}
		case "=":
			if v.Compare(op.version) != 0 {
				return false
			}
		case ">":
			if v.Compare(op.version) <= 0 {
				return false
			}
		case ">=":
			if v.Compare(op.version) < 0 {
				return false
			}
		case "<":
			if v.Compare(op.version) >= 0 {
				return false
			}
		case "<=":
			if v.Compare(op.version) > 0 {
				return false
			}
		case "^":
			// Compatible with: >= v, < next major
			low := op.version
			high := Version{Major: low.Major + 1}
			if v.Compare(low) < 0 || v.Compare(high) >= 0 {
				return false
			}
		case "~":
			// Approximately equivalent: >= v, < next minor
			low := op.version
			high := Version{Major: low.Major, Minor: low.Minor + 1}
			if v.Compare(low) < 0 || v.Compare(high) >= 0 {
				return false
			}
		}
	}
	return true
}

// String returns the constraint string.
func (c *Constraint) String() string {
	return c.raw
}

// BestMatch selects the newest version that satisfies the constraint.
func BestMatch(constraint *Constraint, versions []Version) (Version, bool) {
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Compare(versions[j]) > 0
	})
	for _, v := range versions {
		if constraint.Matches(v) {
			return v, true
		}
	}
	return Version{}, false
}

// MustParseVersion panics on error.
func MustParseVersion(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// Coerce parses a loose version (e.g., "v1.2" -> "1.2.0").
func Coerce(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return Version{}, fmt.Errorf("too many version parts: %q", s)
	}
	nums := make([]int, 3)
	for i, p := range parts {
		// Remove pre-release/build suffix from last numeric part
		if i == len(parts)-1 {
			if idx := strings.IndexAny(p, "-+"); idx >= 0 {
				p = p[:idx]
			}
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version part %q: %w", p, err)
		}
		nums[i] = n
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2]}, nil
}

// FindLatest returns the highest version from a list.
// If the list is empty, returns Version{0,0,0,"",""}.
func FindLatest(versions []Version) Version {
	if len(versions) == 0 {
		return Version{}
	}
	best := versions[0]
	for _, v := range versions[1:] {
		if v.Compare(best) > 0 {
			best = v
		}
	}
	return best
}

// VersionsFromStrings parses a list of version strings, ignoring invalid ones.
func VersionsFromStrings(strs []string) []Version {
	result := make([]Version, 0, len(strs))
	for _, s := range strs {
		v, err := ParseVersion(s)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}
