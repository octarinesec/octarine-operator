package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	versionA Version
	versionB Version
}

func TestIsLargerThanTrue(t *testing.T) {
	testCases := []testCase{
		{versionA: Version("1.0.0"), versionB: Version("0.9.9")},                 // Basic comparison
		{versionA: Version("1.1.1"), versionB: Version("1.1")},                   // Different versions length
		{versionA: Version("1.0.1"), versionB: Version("1.0.0")},                 // Patch increment
		{versionA: Version("1.1.0"), versionB: Version("1.0.10")},                // Minor increment
		{versionA: Version("2.0.0"), versionB: Version("1.9.9")},                 // Major increment
		{versionA: Version("2.0.0-beta"), versionB: Version("2.0.0-alpha")},      // Pre-release comparison
		{versionA: Version("2.0.0"), versionB: Version("2.0.0-beta")},            // Stable vs pre-release
		{versionA: Version("1.2.0-alpha.2"), versionB: Version("1.2.0-alpha.1")}, // Pre-release with multiple segments (up to 9)
		{versionA: Version("1.2.0-beta.1"), versionB: Version("1.2.0-alpha.10")}, // Different pre-release labels
		{versionA: Version("1.11.0"), versionB: Version("1.2.0")},                // Double digits
		{versionA: Version("1.0.10"), versionB: Version("1.0.2")},                // Patch number double digits
		{versionA: Version("11.0.0"), versionB: Version("2.10.0")},               // Major double digits
		{versionA: Version("1.1.1"), versionB: Version("1.1.0")},                 // All segments incremented
		{versionA: Version("3.0.0"), versionB: Version("2.10.10")},               // All previous segments at max
		{versionA: Version("1.0.0-rc.1"), versionB: Version("1.0.0-beta.9")},     // Pre-release progression
		{versionA: Version("1.0.0-rc.3"), versionB: Version("1.0.0-rc.1")},       // Pre-release with numbers (up to 9)
		{versionA: Version("2.0.0"), versionB: Version("2.0.0-rc.10")},           // Stable vs long pre-release
		{versionA: Version("1.0.1"), versionB: Version("1.0.0-alpha")},           // Patch increment over pre-release
		{versionA: Version("2.0.0-beta"), versionB: Version("2.0.0-alpha.10")},   // Pre-release without numbers vs one with numbers
		{versionA: Version("3.2.1"), versionB: Version("3.2.0")},                 // Minor and patch increment
		{versionA: Version("4.0.1"), versionB: Version("4.0.0")},                 // Major increment with patch
		{versionA: Version("5.2.0"), versionB: Version("5.1.10")},                // Major stable with minor and patch increments
		{versionA: Version("1.2.3"), versionB: Version("1.2.0-alpha.4")},         // Stable vs pre-release with multiple segments
		{versionA: Version("1.2.3-beta.5"), versionB: Version("1.2.3-alpha.5")},  // Same version, different pre-release
		{versionA: Version("1.2.3-beta.9"), versionB: Version("1.2.3-beta.2")},   // Same version, pre-release with numbers (up to 9)
		{versionA: Version("10.10.10"), versionB: Version("2.10.10")},            // Large major increment
		{versionA: Version("0.0.1"), versionB: VersionUnknown},                   // Version is always larger than an unknown version
		{versionA: Version("0.0.1"), versionB: MinVersionNone},                   // Version is always larger than a none version
		{versionA: MaxVersionLatest, versionB: Version("0.0.1")},                 // Latest version is always larger than other version
		{versionA: MaxVersionMain, versionB: Version("0.0.1")},                   // Main version is always larger than other version
	}

	testIsLargerThan(t, testCases, true)
}

func TestIsLargerThanFalse(t *testing.T) {
	testCases := []testCase{
		{versionA: Version("0.9.9"), versionB: Version("0.9.9")},                 // Same versions
		{versionA: Version("1.1"), versionB: Version("1.1.1")},                   // Different versions length
		{versionA: MinVersionNone, versionB: MinVersionNone},                     // Same version none
		{versionA: VersionUnknown, versionB: VersionUnknown},                     // Same version unknown
		{versionA: MaxVersionLatest, versionB: MaxVersionLatest},                 // Same version latest
		{versionA: MaxVersionMain, versionB: MaxVersionMain},                     // Same version main
		{versionA: MaxVersionMain, versionB: MaxVersionLatest},                   // Both max versions
		{versionA: Version("1.1"), versionB: Version("1.1.0")},                   // Same version with postfix 0
		{versionA: Version("1.1.0"), versionB: Version("1.1")},                   // Same version with postfix 0
		{versionA: Version("0.9.9"), versionB: Version("1.0.0")},                 // Basic comparison
		{versionA: Version("1.0.0"), versionB: Version("1.0.1")},                 // Patch increment
		{versionA: Version("1.0.10"), versionB: Version("1.1.0")},                // Minor increment
		{versionA: Version("1.9.9"), versionB: Version("2.0.0")},                 // Major increment
		{versionA: Version("2.0.0-alpha"), versionB: Version("2.0.0-beta")},      // Pre-release comparison
		{versionA: Version("2.0.0-beta"), versionB: Version("2.0.0")},            // Stable vs pre-release
		{versionA: Version("1.2.0-alpha.1"), versionB: Version("1.2.0-alpha.2")}, // Pre-release with multiple segments (up to 9)
		{versionA: Version("1.2.0-alpha.10"), versionB: Version("1.2.0-beta.1")}, // Different pre-release labels
		{versionA: Version("1.2.0"), versionB: Version("1.11.0")},                // Double digits
		{versionA: Version("1.0.2"), versionB: Version("1.0.10")},                // Patch number double digits
		{versionA: Version("2.10.0"), versionB: Version("11.0.0")},               // Major double digits
		{versionA: Version("1.1.0"), versionB: Version("1.1.1")},                 // All segments incremented
		{versionA: Version("2.10.10"), versionB: Version("3.0.0")},               // All previous segments at max
		{versionA: Version("1.0.0-beta.9"), versionB: Version("1.0.0-rc.1")},     // Pre-release progression
		{versionA: Version("1.0.0-rc.1"), versionB: Version("1.0.0-rc.3")},       // Pre-release with numbers (up to 9)
		{versionA: Version("2.0.0-rc.10"), versionB: Version("2.0.0")},           // Stable vs long pre-release
		{versionA: Version("1.0.0-alpha"), versionB: Version("1.0.1")},           // Patch increment over pre-release
		{versionA: Version("2.0.0-alpha.10"), versionB: Version("2.0.0-beta")},   // Pre-release without numbers vs one with numbers
		{versionA: Version("3.2.0"), versionB: Version("3.2.1")},                 // Minor and patch increment
		{versionA: Version("4.0.0"), versionB: Version("4.0.1")},                 // Major increment with patch
		{versionA: Version("5.1.10"), versionB: Version("5.2.0")},                // Major stable with minor and patch increments
		{versionA: Version("1.2.0-alpha.4"), versionB: Version("1.2.3")},         // Stable vs pre-release with multiple segments
		{versionA: Version("1.2.3-alpha.5"), versionB: Version("1.2.3-beta.5")},  // Same version, different pre-release
		{versionA: Version("1.2.3-beta.2"), versionB: Version("1.2.3-beta.9")},   // Same version, pre-release with numbers (up to 9)
		{versionA: Version("2.10.10"), versionB: Version("10.10.10")},            // Large major increment
		{versionA: VersionUnknown, versionB: Version("0.0.1")},                   // Version is always larger than an unknown version
		{versionA: MinVersionNone, versionB: Version("0.0.1")},                   // Version is always larger than a none version
		{versionA: Version("0.0.1"), versionB: MaxVersionLatest},                 // Version latest is always larger than any other version
		{versionA: Version("0.0.1"), versionB: MaxVersionMain},                   // Version main is always larger than any other version
		{versionA: Version("0.a.1"), versionB: Version("a.1.1")},                 // Invalid versions considered as equals
	}

	testIsLargerThan(t, testCases, false)
}

func TestIsLessThanTrue(t *testing.T) {
	testCases := []testCase{
		{versionA: Version("0.9.9"), versionB: Version("1.0.0")},                 // Basic comparison
		{versionA: Version("1.1"), versionB: Version("1.1.1")},                   // Different versions length
		{versionA: Version("1.0.0"), versionB: Version("1.0.1")},                 // Patch increment
		{versionA: Version("1.0.10"), versionB: Version("1.1.0")},                // Minor increment
		{versionA: Version("1.9.9"), versionB: Version("2.0.0")},                 // Major increment
		{versionA: Version("2.0.0-alpha"), versionB: Version("2.0.0-beta")},      // Pre-release comparison
		{versionA: Version("2.0.0-beta"), versionB: Version("2.0.0")},            // Stable vs pre-release
		{versionA: Version("1.2.0-alpha.1"), versionB: Version("1.2.0-alpha.2")}, // Pre-release with multiple segments (up to 9)
		{versionA: Version("1.2.0-alpha.10"), versionB: Version("1.2.0-beta.1")}, // Different pre-release labels
		{versionA: Version("1.2.0"), versionB: Version("1.11.0")},                // Double digits
		{versionA: Version("1.0.2"), versionB: Version("1.0.10")},                // Patch number double digits
		{versionA: Version("2.10.0"), versionB: Version("11.0.0")},               // Major double digits
		{versionA: Version("1.1.0"), versionB: Version("1.1.1")},                 // All segments incremented
		{versionA: Version("2.10.10"), versionB: Version("3.0.0")},               // All previous segments at max
		{versionA: Version("1.0.0-beta.9"), versionB: Version("1.0.0-rc.1")},     // Pre-release progression
		{versionA: Version("1.0.0-rc.1"), versionB: Version("1.0.0-rc.3")},       // Pre-release with numbers (up to 9)
		{versionA: Version("2.0.0-rc.10"), versionB: Version("2.0.0")},           // Stable vs long pre-release
		{versionA: Version("1.0.0-alpha"), versionB: Version("1.0.1")},           // Patch increment over pre-release
		{versionA: Version("2.0.0-alpha.10"), versionB: Version("2.0.0-beta")},   // Pre-release without numbers vs one with numbers
		{versionA: Version("3.2.0"), versionB: Version("3.2.1")},                 // Minor and patch increment
		{versionA: Version("4.0.0"), versionB: Version("4.0.1")},                 // Major increment with patch
		{versionA: Version("5.1.10"), versionB: Version("5.2.0")},                // Major stable with minor and patch increments
		{versionA: Version("1.2.0-alpha.4"), versionB: Version("1.2.3")},         // Stable vs pre-release with multiple segments
		{versionA: Version("1.2.3-alpha.5"), versionB: Version("1.2.3-beta.5")},  // Same version, different pre-release
		{versionA: Version("1.2.3-beta.2"), versionB: Version("1.2.3-beta.9")},   // Same version, pre-release with numbers (up to 9)
		{versionA: Version("2.10.10"), versionB: Version("10.10.10")},            // Large major increment
		{versionA: MinVersionNone, versionB: Version("0.0.1")},                   // Version is always larger than a none version
		{versionA: Version("0.0.1"), versionB: MaxVersionLatest},                 // Version latest is always larger than any other version
		{versionA: Version("0.0.1"), versionB: MaxVersionMain},                   // Version main is always larger than any other version
	}

	testIsLessThan(t, testCases, true)
}

func TestIsLessThanFalse(t *testing.T) {
	testCases := []testCase{
		{versionA: Version("0.9.9"), versionB: Version("0.9.9")},                 // Same versions
		{versionA: MinVersionNone, versionB: MinVersionNone},                     // Same version none
		{versionA: VersionUnknown, versionB: VersionUnknown},                     // Same version unknown
		{versionA: MaxVersionLatest, versionB: MaxVersionLatest},                 // Same version latest
		{versionA: MaxVersionMain, versionB: MaxVersionMain},                     // Same version main
		{versionA: MaxVersionMain, versionB: MaxVersionLatest},                   // Both max versions
		{versionA: Version("1.1"), versionB: Version("1.1.0")},                   // Same version with postfix 0
		{versionA: Version("1.1.0"), versionB: Version("1.1")},                   // Same version with postfix 0
		{versionA: Version("1.0.0"), versionB: Version("0.9.9")},                 // Basic comparison
		{versionA: Version("1.1.1"), versionB: Version("1.1")},                   // Different versions length
		{versionA: Version("1.0.1"), versionB: Version("1.0.0")},                 // Patch increment
		{versionA: Version("1.1.0"), versionB: Version("1.0.10")},                // Minor increment
		{versionA: Version("2.0.0"), versionB: Version("1.9.9")},                 // Major increment
		{versionA: Version("2.0.0-beta"), versionB: Version("2.0.0-alpha")},      // Pre-release comparison
		{versionA: Version("2.0.0"), versionB: Version("2.0.0-beta")},            // Stable vs pre-release
		{versionA: Version("1.2.0-alpha.2"), versionB: Version("1.2.0-alpha.1")}, // Pre-release with multiple segments (up to 9)
		{versionA: Version("1.2.0-beta.1"), versionB: Version("1.2.0-alpha.10")}, // Different pre-release labels
		{versionA: Version("1.11.0"), versionB: Version("1.2.0")},                // Double digits
		{versionA: Version("1.0.10"), versionB: Version("1.0.2")},                // Patch number double digits
		{versionA: Version("11.0.0"), versionB: Version("2.10.0")},               // Major double digits
		{versionA: Version("1.1.1"), versionB: Version("1.1.0")},                 // All segments incremented
		{versionA: Version("3.0.0"), versionB: Version("2.10.10")},               // All previous segments at max
		{versionA: Version("1.0.0-rc.1"), versionB: Version("1.0.0-beta.9")},     // Pre-release progression
		{versionA: Version("1.0.0-rc.3"), versionB: Version("1.0.0-rc.1")},       // Pre-release with numbers (up to 9)
		{versionA: Version("2.0.0"), versionB: Version("2.0.0-rc.10")},           // Stable vs long pre-release
		{versionA: Version("1.0.1"), versionB: Version("1.0.0-alpha")},           // Patch increment over pre-release
		{versionA: Version("2.0.0-beta"), versionB: Version("2.0.0-alpha.10")},   // Pre-release without numbers vs one with numbers
		{versionA: Version("3.2.1"), versionB: Version("3.2.0")},                 // Minor and patch increment
		{versionA: Version("4.0.1"), versionB: Version("4.0.0")},                 // Major increment with patch
		{versionA: Version("5.2.0"), versionB: Version("5.1.10")},                // Major stable with minor and patch increments
		{versionA: Version("1.2.3"), versionB: Version("1.2.0-alpha.4")},         // Stable vs pre-release with multiple segments
		{versionA: Version("1.2.3-beta.5"), versionB: Version("1.2.3-alpha.5")},  // Same version, different pre-release
		{versionA: Version("1.2.3-beta.9"), versionB: Version("1.2.3-beta.2")},   // Same version, pre-release with numbers (up to 9)
		{versionA: Version("10.10.10"), versionB: Version("2.10.10")},            // Large major increment
		{versionA: VersionUnknown, versionB: Version("0.0.1")},                   // Version unknown is not considered less than other version
		{versionA: MaxVersionMain, versionB: Version("0.0.1")},                   // Version main is not less than other version
		{versionA: Version("0.0.1"), versionB: VersionUnknown},                   // Version is always larger than an unknown version
		{versionA: Version("0.0.1"), versionB: MinVersionNone},                   // Version is always larger than a none version
		{versionA: Version("0.a.1"), versionB: Version("a.1.1")},                 // Invalid versions considered as equals
	}

	testIsLessThan(t, testCases, false)
}

func testIsLargerThan(t *testing.T, testCases []testCase, expected bool) {
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("testIsLargerThan %d", i), func(t *testing.T) {
			require.Equal(t, expected, testCase.versionA.IsLargerThan(testCase.versionB))
		})
	}
}

func testIsLessThan(t *testing.T, testCases []testCase, expected bool) {
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("testIsLargerThan %d", i), func(t *testing.T) {
			require.Equal(t, expected, testCase.versionA.IsLessThan(testCase.versionB))
		})
	}
}
