package main

import (
	"context"
	"os"
	"time"

	twitterscraper "github.com/n0madic/twitter-scraper"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout)

	scraper := twitterscraper.New()
	err := scraper.Login("BorisovChef2006", "110098867")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to login")
	}

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for tweet := range scraper.SearchTweets(c, "$SOL", 50) {
		if tweet.Error != nil {
			logger.Error().Err(tweet.Error).Msg("Failed to scrape tweet")
		} else {
			logger.Info().Msg(tweet.Text)
		}
	}
}
