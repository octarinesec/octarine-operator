package applyment

func MutateInt64(desiredValue int64, get func() *int64, set func(int64)) bool {
	actualValue := get()
	if actualValue == nil || *actualValue != desiredValue {
		set(desiredValue)
		return true
	}

	return false
}

func MutateInt32(desiredValue int32, get func() *int32, set func(int32)) bool {
	actualValue := get()
	if actualValue == nil || *actualValue != desiredValue {
		set(desiredValue)
		return true
	}

	return false
}

func MutateBool(desiredValue bool, get func() *bool, set func(bool)) bool {
	actualValue := get()
	if actualValue == nil || *actualValue != desiredValue {
		set(desiredValue)
		return true
	}

	return false
}

func MutateString(desiredValue string, get func() *string, set func(string)) bool {
	actualValue := get()
	if actualValue == nil || *actualValue != desiredValue {
		set(desiredValue)
		return true
	}

	return false
}

func MutateStringsMap(desiredMap map[string]string, get func() map[string]string, set func(map[string]string)) bool {
	actualMap := get()
	if actualMap == nil {
		set(desiredMap)
		return true
	}

	mutated := false
	for key, desiredValue := range desiredMap {
		actualValue, ok := desiredMap[key]
		if !ok || actualValue != desiredValue {
			actualMap[key] = desiredValue
			mutated = true
		}
	}

	return mutated
}
