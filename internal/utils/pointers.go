package utils

// StringPtr takes a string value and returns a pointer to that string.
func StringPtr(s string) *string {
	return &s
}

// IntPtr takes an integer value and returns a pointer to that integer.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr takes a bool value and returns a pointer to that bool.
func BoolPtr(b bool) *bool {
	return &b
}

// Float64Ptr takes a float64 value and returns a pointer to that float64.
func Float64Ptr(f float64) *float64 {
	return &f
}

// Float32Ptr takes a float32 value and returns a pointer to that float32.
func Float32Ptr(f float32) *float32 {
	return &f
}
