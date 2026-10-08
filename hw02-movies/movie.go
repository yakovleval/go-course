package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

const movieURLPattern string = "https://homeworksite.site/%d/info.0.json"

func movieURL(id int) string {
	return fmt.Sprintf(movieURLPattern, id)
}

func fetchMovie(ctx context.Context, timeout time.Duration, client *http.Client, id int) (Movie, error) {
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, movieURL(id), nil)
	if err != nil {
		return Movie{}, fmt.Errorf("create request for movie %d: %w", id, err)
	}
	resp, err := client.Do(request)
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
