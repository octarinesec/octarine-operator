package applyment

func EnforceMapContains(actualMap map[string]string, desiredMap map[string]string) {
	for desiredKey, desiredValue := range desiredMap {
		actualMap[desiredKey] = desiredValue
	}
}
