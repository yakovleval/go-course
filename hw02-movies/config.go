package main

import (
	"flag"
	"fmt"
	"time"
)

type Config struct {
	From    int
	To      int
	Workers int
	Timeout time.Duration
}

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
