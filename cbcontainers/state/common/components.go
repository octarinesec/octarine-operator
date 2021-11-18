package common

func IsEnabled(componentEnabledField *bool) bool {
	return componentEnabledField != nil && *componentEnabledField
}

func IsDisabled(componentEnabledField *bool) bool {
	return !IsEnabled(componentEnabledField)
}
