package primes

import (
	"fmt"
	"runtime"
	"testing"
)

func TestIsPrime(t *testing.T) {
	cases := []struct {
		input int
		want  bool
	}{
		{0, false},
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{17, true},
		{18, false},
		{19, true},
		{20, false},
	}

	for _, c := range cases {
		got := isPrime(c.input)
		if got != c.want {
			t.Errorf("isPrime(%d) = %v; want %v", c.input, got, c.want)
		}
	}
}

func TestFindPrimes(t *testing.T) {
	cases := []struct {
		start, end, want int
	}{
		{0, 10, 4},  // primes: 2,3,5,7
		{10, 20, 4}, // primes: 11,13,17,19
		{0, 1, 0},   // no primes
		{20, 10, 0}, // invalid range
		{-10, 2, 1}, // primes: 2
		{2, 2, 1},   // single prime
	}

	for _, c := range cases {
		got := FindPrimes(c.start, c.end)
		if got != c.want {
			t.Errorf("FindPrimes(%d, %d) = %d; want %d", c.start, c.end, got, c.want)
		}
	}
}

func TestFindPrimesGoroutines(t *testing.T) {
	cases := []struct {
		start, end, numGoroutines int
	}{
		{0, 10, 2},
		{10, 20, 4},
		{0, 1, 1},
		{20, 10, 3},
		{-10, 2, 2},
		{2, 2, 1},
	}

	for _, c := range cases {
		want := FindPrimes(c.start, c.end)
		got := FindPrimesGoroutines(c.start, c.end, c.numGoroutines)
		if got != want {
			t.Errorf("FindPrimesGoroutines(%d, %d, %d) = %d; want %d", c.start, c.end, c.numGoroutines, got, want)
		}
	}
}

func BenchmarkFindPrimes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FindPrimes(0, 1000000)
	}
}

func BenchmarkFindPrimesGoroutines(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FindPrimesGoroutines(0, 1000000, 4)
	}
}

func BenchmarkPrimes(b *testing.B) {
	rangeSizes := []int{10_000, 100_000, 1_000_000}
	numGoroutines := []int{1, 2, 4, 8, runtime.NumCPU()}

	for _, size := range rangeSizes {
		// Sequential benchmark
		b.Run(fmt.Sprintf("Sequential_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				FindPrimes(0, size)
			}
		})

		// Concurrent benchmarks
		for _, g := range numGoroutines {
			b.Run(fmt.Sprintf("Goroutines_%d_%d", size, g), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					FindPrimesGoroutines(0, size, g)
				}
			})
		}
	}
}
