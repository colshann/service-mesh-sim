package primes

// isPrime checks if a number is prime using trial division up to the square root of the number.
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// FindPrimes returns the number of prime numbers from start to end.
func FindPrimes(start int, end int) int {
	count := 0
	for i := start; i <= end; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

// FindPrimesGoroutines returns the number of prime numbers from start to end using goroutines for concurrent processing. Uses chunking for division of work among goroutines.
func FindPrimesGoroutines(start int, end int, numGoroutines int) int {
	count := 0
	chunkSize := (end - start) / numGoroutines
	ch := make(chan int, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		s := start + i*chunkSize
		e := s + chunkSize - 1
		if i == numGoroutines-1 {
			e = end
		}

		go func(s, e int) {
			localCount := 0
			for j := s; j <= e; j++ {
				if isPrime(j) {
					localCount++
				}
			}
			ch <- localCount
		}(s, e)
	}
	for i := 0; i < numGoroutines; i++ {
		count += <-ch
	}
	return count
}
