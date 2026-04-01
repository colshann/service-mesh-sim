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

// FindPrimes returns the number of prime numbers from 1 to n.
func FindPrimes(n int) int {
	count := 0
	for i := 2; i <= n; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

// FindPrimesGoroutines returns the number of prime numbers from 1 to n using goroutines for concurrent processing.
func FindPrimesGoroutines(n int, numGoroutines int) int {
	count := 0
	ch := make(chan int)
	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			localCount := 0
			for j := start; j <= n; j += numGoroutines {
				if isPrime(j) {
					localCount++
				}
			}
			ch <- localCount
		}(i + 2)
	}
	for i := 0; i < numGoroutines; i++ {
		count += <-ch
	}
	return count
}
