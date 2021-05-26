package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	maxGoroutines chan int
	wg            sync.WaitGroup
)

func doCalculation(values []int64, c chan []int64) {
	result := make([]int64, 0)

	for _, value := range values {
		prime := value != 0 // zero is not a prime number...
		for div := int64(2); div < value; div++ {
			if value%div == 0 && value > 0 {
				prime = false
				break
			}
		}

		if prime {
			result = append(result, value)
		}
	}

	// free the channel for more goroutines and the waitgroup.
	<-maxGoroutines

	c <- result
}

func primes(values []int64, sliceSize uint, c chan []int64) {

	result := make([]int64, 0)
	subChain := make(chan []int64)
	iterationCount := 0

	// func to drain the result channel
	go func() {
		for {
			subResult := <-subChain
			result = append(result, subResult...)
			wg.Done()
		}
	}()

	for count := uint(len(values)) / sliceSize; count > 0; count-- {
		sliceEnd := count * sliceSize
		sliceInit := sliceEnd - sliceSize

		// add one goroutine to channel for concurency control, and also add 1 to waitGroup
		maxGoroutines <- 1
		wg.Add(1)

		go doCalculation(values[sliceInit:sliceEnd], subChain)
		iterationCount++
		fmt.Printf("\nsplit %d", iterationCount)
	}

	// wait for all primes...
	wg.Wait()

	fmt.Printf("\nSearch Finalized sorting result....")
	sort.Slice(result, func(a, b int) bool {
		return result[a] < result[b]
	})

	c <- result

}

func main() {

	var (
		maxPrime         uint64 = 100000
		sliceSize        uint64 = 10000
		err              error
		maxGoroutinesInt int64 = 0
	)

	if len(os.Args) > 1 {
		if maxPrime, err = strconv.ParseUint(os.Args[1], 10, 64); err != nil {
			log.Panicf("maxPrime is not valid, need to be a positive integer (64 bits long), value passed %#v", os.Args[1])
		}

		if len(os.Args) > 2 {
			if sliceSize, err = strconv.ParseUint(os.Args[2], 10, 32); err != nil {
				log.Panicf("sliceSize is not valid, need to be a positive integer (32 bits long), value passed %#v", os.Args[2])
			}
		}

		if len(os.Args) > 3 {
			if maxGoroutinesInt, err = strconv.ParseInt(os.Args[3], 10, 32); err != nil {
				log.Panicf("maxGoroutine is not valid, need to be a positive integer, value passed %v", os.Args[3])
			}
		}
	}

	if maxGoroutinesInt == 0 { // 0 = use number os cpu/threads
		threads := runtime.NumCPU()
		maxGoroutines = make(chan int, threads) // get the actual number os core/threads on cpu
		fmt.Printf("Using the core/threads of cpu %d to maxGoroutines\n", threads)
	} else {
		maxGoroutines = make(chan int, maxGoroutinesInt)
		fmt.Printf("Using the passed value of maxGoroutines %d\n", maxGoroutinesInt)
	}

	fmt.Printf("Search for primes 0 to %d, on slices of %d\n", maxPrime, sliceSize)

	c := make(chan []int64)

	values := make([]int64, maxPrime)
	for index := range values {
		values[index] = int64(index)
	}

	fmt.Printf("Initializing search for %d primes\n", len(values))
	initTime := time.Now()
	go primes(values, uint(sliceSize), c)

	result := <-c

	fmt.Printf("%#v\n", result)

	fmt.Printf("\n\nElapsed %dms", time.Since(initTime).Milliseconds())

}
