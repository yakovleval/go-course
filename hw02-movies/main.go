package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	client := &http.Client{}
	jobs := make(chan int)
	results := make(chan Result)
	var wg sync.WaitGroup
	startWorkers(ctx, config.Workers, config.Timeout, client, jobs, results, &wg)
	go produceJobs(ctx, config.From, config.To, jobs)
	go closeResultsAfterWorkers(&wg, results)
	for result := range results {
		if result.Err != nil {
			fmt.Fprintln(os.Stderr, result.Err)
			continue
		}
		fmt.Printf("%d — %s — %d — %s\n", result.Movie.ID, result.Movie.Title, result.Movie.Year, result.Movie.Director)
	}
}
