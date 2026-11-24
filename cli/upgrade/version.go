package upgrade

import (
	"fmt"
	"strconv"
	"strings"
)

// Semver represents a parsed major.minor.patch version.
type Semver struct {
	Major int
	Minor int
	Patch int
}

// Normalize trims whitespace and any leading "v".
func Normalize(raw string) string {
	v := strings.TrimSpace(raw)
	return strings.TrimPrefix(v, "v")
}

// Parse parses a strict X.Y.Z version (with optional leading "v").
func Parse(raw string) (Semver, error) {
	v := Normalize(raw)
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return Semver{}, fmt.Errorf("invalid version %q", raw)
	}
	var nums [3]int
	for i, part := range parts {
		if part == "" {
			return Semver{}, fmt.Errorf("invalid version %q", raw)
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return Semver{}, fmt.Errorf("invalid version %q", raw)
		}
		nums[i] = n
	}
	return Semver{Major: nums[0], Minor: nums[1], Patch: nums[2]}, nil
}

// Compare returns -1 if a < b, 0 if a == b, and 1 if a > b.
func Compare(a, b string) (int, error) {
	va, err := Parse(a)
	if err != nil {
		return 0, err
	}
	vb, err := Parse(b)
	if err != nil {
		return 0, err
	}
	if va.Major != vb.Major {
		if va.Major < vb.Major {
			return -1, nil
		}
		return 1, nil
	}
	if va.Minor != vb.Minor {
		if va.Minor < vb.Minor {
			return -1, nil
		}
		return 1, nil
	}
	if va.Patch != vb.Patch {
		if va.Patch < vb.Patch {
			return -1, nil
		}
		return 1, nil
	}
	return 0, nil
}

// IsNewer reports whether latest is strictly greater than current.
// Returns false if either string is not a valid semver (for example "dev").
func IsNewer(current, latest string) bool {
	cmp, err := Compare(current, latest)
	return err == nil && cmp < 0
}
