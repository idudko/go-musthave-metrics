package handler

// float64Ptr returns a pointer to the given float64 value.
// This is a utility function for tests that need to create pointers to float64 values.
func float64Ptr(v float64) *float64 {
	return &v
}

// int64Ptr returns a pointer to the given int64 value.
// This is a utility function for tests that need to create pointers to int64 values.
func int64Ptr(v int64) *int64 {
	return &v
}
