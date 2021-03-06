package main

import (
	"context"
	"log"
	"strings"

	"github.com/PratikMahajan/Twitter-Knative-Serverless-App-Source/config"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	consumerKey    = strings.TrimSpace(config.MustGetEnvVar("T_CONSUMER_KEY", ""))
	consumerSecret = strings.TrimSpace(config.MustGetEnvVar("T_CONSUMER_SECRET", ""))
	accessToken    = strings.TrimSpace(config.MustGetEnvVar("T_ACCESS_TOKEN", ""))
	accessSecret   = strings.TrimSpace(config.MustGetEnvVar("T_ACCESS_SECRET", ""))
)

func search(ctx context.Context, query, sink string, stop <-chan struct{}) {

	// twitter client config
	cfg := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := cfg.Client(oauth1.NoContext, token)
	twClient := twitter.NewClient(httpClient)

	sinker, err := newSinkPoster(sink)
	if err != nil {
		log.Fatalf("Error getting sinker: %v", err)
	}

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(t *twitter.Tweet) {
		log.Printf("Got tweet: %s\n", t.IDStr)
		if err := sinker.post(ctx, t); err != nil {
			log.Printf("Error on tweet handle: %v\n", err)
		}
	}

	params := &twitter.StreamFilterParams{
		Track:         []string{query},
		StallWarnings: twitter.Bool(true),
		Language:      []string{"en"},
	}

	stream, err := twClient.Streams.Filter(params)
	if err != nil {
		log.Fatalf("Error while creating filter: %v\n", err)
		return
	}

	log.Printf("Starting tweet streamming for: %s\n", query)
	go demux.HandleChan(stream.Messages)

}