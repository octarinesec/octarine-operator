package applyment

func EnforceMapContains(actualMap map[string]string, desiredMap map[string]string) {
	for desiredKey, desiredValue := range desiredMap {
		actualMap[desiredKey] = desiredValue
	}
}

func RemoveMapContains(actualMap map[string]string, desiredMap map[string]string) {
	for desiredKey := range desiredMap {
		if _, fieldExists := actualMap[desiredKey]; fieldExists {
			delete(actualMap, desiredKey)
		}
	}
}