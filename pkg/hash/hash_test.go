package hash

import (
	"testing"
)

// Benchmarks
func BenchmarkComputeHash_Small(b *testing.B) {
	key := "secret-key"
	data := []byte("small data")

	b.ResetTimer()
	for b.Loop() {
		ComputeHash(data, key)
	}
}

func BenchmarkComputeHash_Medium(b *testing.B) {
	key := "secret-key"
	data := make([]byte, 1024) // 1KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for b.Loop() {
		ComputeHash(data, key)
	}
}

func BenchmarkComputeHash_Large(b *testing.B) {
	key := "secret-key"
	data := make([]byte, 10240) // 10KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for b.Loop() {
		ComputeHash(data, key)
	}
}

func BenchmarkValidateHash_Valid(b *testing.B) {
	key := "secret-key"
	data := []byte("test data")
	validHash := ComputeHash(data, key)

	b.ResetTimer()
	for b.Loop() {
		ValidateHash(data, key, validHash)
	}
}

func BenchmarkValidateHash_Invalid(b *testing.B) {
	key := "secret-key"
	data := []byte("test data")
	invalidHash := "invalid_hash_12345"

	b.ResetTimer()
	for b.Loop() {
		ValidateHash(data, key, invalidHash)
	}
}
