package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Config struct {
	From    int
	To      int
	Workers int
	Timeout time.Duration
}

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type Result struct {
	ID    int
	Movie Movie
	Err   error
}

const movieURLPattern string = "https://homeworksite.site/%d/info.0.json"

func parseConfig() (Config, error) {
	from := flag.Int("from", 0, "first movie ID")
	to := flag.Int("to", 0, "last movie ID")
	workers := flag.Int("workers", 10, "number of workers")
	timeout := flag.Duration("timeout", 5*time.Second, "timeout")

	flag.Parse()
	provided := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})
	if !provided["from"] {
		return Config{}, fmt.Errorf("--from is required")
	}
	if !provided["to"] {
		return Config{}, fmt.Errorf("--to is required")
	}
	if *from > *to {
		return Config{}, fmt.Errorf("--from %d must be less than or equal to --to %d", *from, *to)
	}
	if *workers <= 0 {
		return Config{}, fmt.Errorf("--workers %d must be greater than 0", *workers)
	}
	if *timeout <= 0 {
		return Config{}, fmt.Errorf("--timeout %s must be greater than 0", *timeout)
	}
	return Config{
		From:    *from,
		To:      *to,
		Workers: *workers,
		Timeout: *timeout,
	}, nil
}

func fetchMovie(client *http.Client, id int) (Movie, error) {
	resp, err := client.Get(movieURL(id))
	if err != nil {
		return Movie{}, fmt.Errorf("fetch movie %d: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Movie{}, fmt.Errorf("fetch movie %d: %s", id, resp.Status)
	}
	var movie Movie
	var decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&movie)
	if err != nil {
		return Movie{}, fmt.Errorf("decode movie %d: %w", id, err)
	}
	return movie, nil
}

func movieURL(id int) string {
	return fmt.Sprintf(movieURLPattern, id)
}

func worker(client *http.Client, jobs <-chan int, results chan<- Result) {
	for id := range jobs {
		movie, err := fetchMovie(client, id)
		results <- Result{ID: id, Movie: movie, Err: err}
	}
}

func produceJobs(from, to int, jobs chan<- int) {
	defer close(jobs)
	for id := from; id <= to; id++ {
		jobs <- id
	}
}

func startWorkers(
	count int,
	client *http.Client,
	jobs <-chan int,
	results chan<- Result,
	wg *sync.WaitGroup,
) {
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			worker(client, jobs, results)
		}()
	}
}

func closeResultsAfterWorkers(wg *sync.WaitGroup, results chan<- Result) {
	wg.Wait()
	close(results)
}

func main() {
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client := &http.Client{
		Timeout: config.Timeout,
	}
	jobs := make(chan int)
	results := make(chan Result)
	var wg sync.WaitGroup
	startWorkers(config.Workers, client, jobs, results, &wg)
	go produceJobs(config.From, config.To, jobs)
	go closeResultsAfterWorkers(&wg, results)
	for result := range results {
		if result.Err != nil {
			fmt.Fprintln(os.Stderr, result.Err)
			continue
		}
		fmt.Printf("%d — %s — %d — %s\n", result.Movie.ID, result.Movie.Title, result.Movie.Year, result.Movie.Director)
	}
}
