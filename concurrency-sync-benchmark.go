package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type readerCallback func(int, *sync.WaitGroup) int64
type writerCallback func(int, *sync.WaitGroup)

var mutex sync.Mutex
var rwmutex sync.RWMutex
var ch chan int64
var buffCh chan int64
var val int64
var atomicVal atomic.Int64

// Mutexes

func mutexReader(iterations int, wg *sync.WaitGroup) int64 {
	var x int64
	for i := 0; i < iterations; i++ {
		mutex.Lock()
		x = val
		mutex.Unlock()
	}
	wg.Done()
	return x
}

func mutexWriter(iterations int, wg *sync.WaitGroup) {
	for i := 0; i < iterations; i++ {
		mutex.Lock()
		val++
		mutex.Unlock()
	}
	wg.Done()
}

// RWMutexes
func rwMutexReader(iterations int, wg *sync.WaitGroup) int64 {
	var x int64
	for i := 0; i < iterations; i++ {
		rwmutex.RLock()
		x = val
		rwmutex.RUnlock()
	}
	wg.Done()
	return x
}

func rwMutexWriter(iterations int, wg *sync.WaitGroup) {
	for i := 0; i < iterations; i++ {
		rwmutex.Lock()
		val++
		rwmutex.Unlock()
	}
	wg.Done()
}

// Channels

func chanReader(iterations int, wg *sync.WaitGroup) int64 {
	var x int64
	for i := 0; i < iterations; i++ {
		x = <-ch
	}
	wg.Done()
	return x
}

func chanWriter(iterations int, wg *sync.WaitGroup) {
	var i int64
	for i = 0; i < int64(iterations); i++ {
		ch <- i + 1 // we increment because we must be fair to other implementations
	}
	wg.Done()
}

// Buffered Channels

func buffChanReader(iterations int, wg *sync.WaitGroup) int64 {
	var x int64
	for i := 0; i < iterations; i++ {
		x = <-buffCh
	}
	wg.Done()
	return x
}

func buffChanWriter(iterations int, wg *sync.WaitGroup) {
	var i int64
	for i = 0; i < int64(iterations); i++ {
		buffCh <- i + 1 // we increment because we must be fair to other implementations
	}
	wg.Done()
}

// Atomic

func atomicReader(iterations int, wg *sync.WaitGroup) int64 {
	var x int64
	for i := 0; i < iterations; i++ {
		x = atomicVal.Load()
	}
	wg.Done()
	return x
}

func atomicWriter(iterations int, wg *sync.WaitGroup) {
	for i := 0; i < iterations; i++ {
		atomicVal.Add(1)
	}
	wg.Done()
}

// spin workers and measure time

func run(reader readerCallback, writer writerCallback, numWorkers int, iterations int) time.Duration {
	var wg sync.WaitGroup
	wg.Add(numWorkers * 2)
	startTime := time.Now()
	for i := 0; i < numWorkers; i++ {
		go writer(iterations, &wg)
		go reader(iterations, &wg)
	}
	wg.Wait()
	return time.Since(startTime)
}

func printRun(reader readerCallback, writer writerCallback, numWorkers int, numOperations int) {
	iterations := int(numOperations / numWorkers)
	duration := run(reader, writer, numWorkers, iterations)
	fmt.Printf("%.0d\t\t", duration.Nanoseconds()/int64(numOperations))
}

// main

func main() {

	if len(os.Args) < 3 {
		panic("Usage: concurrency-sync-benchmark <millions operations per run> <maximum number of workers>")
	}

	numOperations, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("Number of operations must be integer")
	}
	numOperations = numOperations * 1e6

	maxWorkers, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic("Maximum number of workers must be integer")
	}

	ch = make(chan int64)
	buffCh = make(chan int64, numOperations)

	fmt.Println("Workers\t\tMutex(ns)\tRWMutex(ns)\tChannels(ns)\tBuffChannels\tAtomic(ns)")

	for numWorkers := 1; numWorkers < maxWorkers; numWorkers++ {
		if numOperations%numWorkers != 0 {
			continue
		}
		val = 0
		atomicVal.Store(0)
		fmt.Print(numWorkers, "\t\t")
		printRun(mutexReader, mutexWriter, numWorkers, numOperations)
		printRun(rwMutexReader, rwMutexWriter, numWorkers, numOperations)
		printRun(chanReader, chanWriter, numWorkers, numOperations)
		printRun(buffChanReader, buffChanWriter, numWorkers, numOperations)
		printRun(atomicReader, atomicWriter, numWorkers, numOperations)
		fmt.Println(" ")
	}

}
