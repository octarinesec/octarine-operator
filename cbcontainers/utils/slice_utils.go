package utils

func StringsSlicesHaveSameItems(firstSlice []string, secondSlice []string) bool {
	if len(firstSlice) != len(secondSlice) {
		return false
	}

	sliceValues := make(map[string]int)
	for _, value := range firstSlice {
		sliceValues[value] += 1
	}

	for _, value := range secondSlice {
		if _, ok := sliceValues[value]; !ok {
			return false
		}
		sliceValues[value] -= 1
	}

	for _, occurrences := range sliceValues {
		if occurrences != 0 {
			return false
		}
	}

	return true
}
