package models

import (
	"strconv"
	"strings"
)

type Version string

const (
	MinVersionNone   Version = "none"
	MaxVersionLatest Version = "latest"
	MaxVersionMain   Version = "main"
	VersionUnknown   Version = ""
)

func (v Version) isMaxVersion() bool {
	return v == MaxVersionLatest || v == MaxVersionMain
}

func (v Version) CompareWith(other Version) int {
	// Split into version and pre-release parts
	partsV := strings.SplitN(string(v), "-", 2)
	partsOther := strings.SplitN(string(other), "-", 2)

	numbersV := strings.Split(partsV[0], ".")
	numbersOther := strings.Split(partsOther[0], ".")

	// Normalize lengths: append zeros to the shorter version
	for len(numbersV) < len(numbersOther) {
		numbersV = append(numbersV, "0")
	}
	for len(numbersOther) < len(numbersV) {
		numbersOther = append(numbersOther, "0")
	}

	for i := 0; i < len(numbersV); i++ {
		intA, errA := strconv.Atoi(numbersV[i])
		intB, errB := strconv.Atoi(numbersOther[i])

		if errA != nil || errB != nil {
			return 0
		}

		if intA > intB {
			return 1
		} else if intA < intB {
			return -1
		}
	}

	// If one version has a pre-release and the other doesn't
	if len(partsV) == 1 && len(partsOther) > 1 {
		return 1
	}
	if len(partsV) > 1 && len(partsOther) == 1 {
		return -1
	}

	// If both have pre-releases, compare them lexicographically
	if len(partsV) > 1 && len(partsOther) > 1 {
		if partsV[1] > partsOther[1] {
			return 1
		} else if partsV[1] < partsOther[1] {
			return -1
		}
	}

	return 0
}

func (v Version) IsLargerThan(other Version) bool {
	if v == other {
		return false
	}

	if v.isMaxVersion() && other.isMaxVersion() {
		return false
	}

	if v == MaxVersionLatest || v == MaxVersionMain {
		return true
	}

	if other == MaxVersionLatest || other == MaxVersionMain {
		return false
	}

	if v == MinVersionNone || v == VersionUnknown {
		return false
	}

	if other == MinVersionNone || other == VersionUnknown {
		return true
	}

	return v.CompareWith(other) > 0
}

func (v Version) IsLessThan(other Version) bool {
	if v == other {
		return false
	}

	if v.isMaxVersion() && other.isMaxVersion() {
		return false
	}

	if v == MinVersionNone {
		return true
	}

	if other == MinVersionNone {
		return false
	}

	if v == MaxVersionLatest || v == MaxVersionMain || v == VersionUnknown {
		return false
	}

	if other == VersionUnknown {
		return false
	}

	if other == MaxVersionLatest || other == MaxVersionMain {
		return true
	}

	return v.CompareWith(other) < 0
}
