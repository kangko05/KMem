package tests

import (
	"context"
	"fmt"
	"gateway/internal/utils"
	"gateway/protogen"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestT(t *testing.T) {
	assert := assert.New(t)

	conn, err := grpc.NewClient(utils.FILESERVICE, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(err)
	defer conn.Close()

	log.Println("connection successful")

	client := protogen.NewFileServiceClient(conn)

	log.Println("client build successful")

	stream, err := client.Upload(t.Context())
	assert.Nil(err)

	rb, err := os.ReadFile("/home/kang/Pictures/output2.jpg")
	assert.Nil(err)

	err = stream.Send(&protogen.UploadRequest{Chunk: rb, Filename: "output.jpg"})
	assert.Nil(err)

	reply, err := stream.CloseAndRecv()
	assert.Nil(err)

	log.Println(reply.GetStatus())
	log.Println(reply.GetMsg())
}

func TestQueue(t *testing.T) {
	t.Run("basic queue operations", func(t *testing.T) {
		q := utils.NewQueue[int]()

		// Run queue with worker function that multiplies each item by 10
		q.Run(t.Context(), func(i int) {
			fmt.Println(i * 10)
		})

		// Add 100 items to the queue
		for i := range 100 {
			q.Add(i)
		}

		// Wait for processing to finish
		time.Sleep(5 * time.Second)
	})

	t.Run("batch adding items", func(t *testing.T) {
		q := utils.NewQueue[int]()

		// Create a slice of items to add in batch
		items := make([]int, 100)
		for i := range items {
			items[i] = i * 10
		}

		// Add all items at once
		q.Run(t.Context(), func(i int) {
			fmt.Printf("Processing item: %d\n", i)
		})

		q.Add(items...)

		// Wait for processing to finish
		time.Sleep(3 * time.Second)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		q := utils.NewQueue[int]()

		// Use a wait group to track worker completion
		var wg sync.WaitGroup
		wg.Add(1)

		q.Run(ctx, func(i int) {
			// Simulate work
			time.Sleep(50 * time.Millisecond)
			fmt.Printf("Processed: %d\n", i)
		})

		// Add some items
		for i := range 50 {
			q.Add(i)
		}

		// Cancel after a short time
		go func() {
			time.Sleep(300 * time.Millisecond)
			fmt.Println("Cancelling context...")
			cancel()

			// Give time for workers to shut down
			time.Sleep(500 * time.Millisecond)
			wg.Done()
		}()

		wg.Wait()
		// Should exit gracefully without processing all items
	})

	t.Run("concurrent adding", func(t *testing.T) {
		q := utils.NewQueue[int]()

		// Set up a simple processor
		processed := 0
		var mu sync.Mutex

		q.Run(t.Context(), func(i int) {
			mu.Lock()
			processed++
			mu.Unlock()
		})

		// Start multiple goroutines that add items concurrently
		const numGoroutines = 10
		const itemsPerGoroutine = 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for g := range numGoroutines {
			go func(offset int) {
				defer wg.Done()
				for i := range itemsPerGoroutine {
					q.Add(offset*itemsPerGoroutine + i)
				}
			}(g)
		}

		// Wait for all additions to complete
		wg.Wait()

		// Allow time for processing
		time.Sleep(8 * time.Second)

		// Verify all items were processed
		mu.Lock()
		result := processed
		mu.Unlock()

		expected := numGoroutines * itemsPerGoroutine
		if result != expected {
			t.Errorf("Expected %d processed items, got %d", expected, result)
		} else {
			fmt.Printf("Successfully processed all %d items\n", result)
		}
	})

	t.Run("burst pattern upload", func(t *testing.T) {
		// Simulate "parents' upload pattern" - long idle, then burst of items
		q := utils.NewQueue[string]()

		// Track processed items
		processedItems := make(map[string]bool)
		var mu sync.Mutex

		q.Run(t.Context(), func(item string) {
			// Simulate varying processing times
			processTime := 10 + time.Duration(len(item))*time.Millisecond
			time.Sleep(processTime)

			mu.Lock()
			processedItems[item] = true
			mu.Unlock()
		})

		// Create burst data
		burstSize := 500
		burstData := make([]string, burstSize)
		for i := range burstSize {
			burstData[i] = fmt.Sprintf("photo_%d.jpg", i)
		}

		// Add all items in one burst
		t0 := time.Now()
		q.Add(burstData...)
		addTime := time.Since(t0)
		fmt.Printf("Added %d items in %v\n", burstSize, addTime)

		// Wait for processing
		time.Sleep(5 * time.Second)

		// Verify all items were processed
		mu.Lock()
		if len(processedItems) != burstSize {
			t.Errorf("Expected %d processed items, got %d", burstSize, len(processedItems))
		} else {
			fmt.Printf("Successfully processed all %d items in burst\n", burstSize)
		}
		mu.Unlock()
	})
}
