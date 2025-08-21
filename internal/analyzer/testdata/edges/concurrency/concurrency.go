package concurrency

import (
	"sync"
	"time"
)

// EdgeTypeSpawnsGoroutine = "spawns_goroutine"
// EdgeTypeSendsChannel = "sends_channel"
// EdgeTypeReceivesChannel = "receives_channel"

// Goroutine spawning
func SpawnGoroutines() {
	// Simple goroutine
	go worker(1)
	
	// Goroutine with anonymous function
	go func() {
		println("anonymous goroutine")
	}()
	
	// Goroutine with closure
	msg := "hello"
	go func() {
		println(msg) // captures msg
	}()
	
	// Multiple goroutines
	for i := 0; i < 5; i++ {
		go worker(i)
	}
}

func worker(id int) {
	println("worker", id)
}

// Channel operations
func ChannelOperations() {
	// Create channels
	ch := make(chan int)
	buffered := make(chan string, 10)
	
	// Send to channel
	go func() {
		ch <- 42 // sends to channel
		buffered <- "hello"
	}()
	
	// Receive from channel
	value := <-ch // receives from channel
	msg := <-buffered
	
	println(value, msg)
}

// Producer-consumer pattern
func ProducerConsumer() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	
	// Spawn workers
	for w := 1; w <= 3; w++ {
		go consumerWorker(w, jobs, results)
	}
	
	// Producer
	go func() {
		for j := 1; j <= 5; j++ {
			jobs <- j // sends to channel
		}
		close(jobs)
	}()
	
	// Collect results
	for r := 1; r <= 5; r++ {
		<-results // receives from channel
	}
}

func consumerWorker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs { // receives from channel
		results <- j * 2 // sends to channel
	}
}

// Select statement with channels
func SelectChannels(ch1, ch2 <-chan string, quit <-chan bool) {
	for {
		select {
		case msg1 := <-ch1: // receives from channel
			println("ch1:", msg1)
		case msg2 := <-ch2: // receives from channel
			println("ch2:", msg2)
		case <-quit: // receives from channel
			return
		case <-time.After(1 * time.Second):
			println("timeout")
		}
	}
}

type Service struct {
	name string
}

func (s *Service) Process() {
	// Service process method
}

// Method spawning goroutines
func (s *Service) StartAsync() {
	go s.Process() // spawns goroutine calling method
	
	// With wait group
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Process()
	}()
	wg.Wait()
}