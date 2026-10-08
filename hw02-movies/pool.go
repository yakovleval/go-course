package main

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	ID    int
	Movie Movie
	Err   error
}

func worker(ctx context.Context, client *http.Client, timeout time.Duration, jobs <-chan int, results chan<- Result) {
	for id := range jobs {
		movie, err := fetchMovie(ctx, timeout, client, id)
		select {
		case <-ctx.Done():
			return
		case results <- Result{ID: id, Movie: movie, Err: err}:
		}
	}
}

func produceJobs(ctx context.Context, from, to int, jobs chan<- int) {
	defer close(jobs)
	for id := from; id <= to; id++ {
		select {
		case <-ctx.Done():
			return
		case jobs <- id:
		}
	}
}

func startWorkers(
	ctx context.Context,
	count int,
	timeout time.Duration,
	client *http.Client,
	jobs <-chan int,
	results chan<- Result,
	wg *sync.WaitGroup,
) {
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			worker(ctx, client, timeout, jobs, results)
		}()
	}
}

func closeResultsAfterWorkers(wg *sync.WaitGroup, results chan<- Result) {
	wg.Wait()
	close(results)
}
