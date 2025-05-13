package main

import (
	"fmt"
	"time"

	"github.com/muun/recovery/electrum"
	"github.com/muun/recovery/survey"
)

func main() {
	config := &survey.Config{
		InitialServers:     electrum.PublicServers,
		Workers:            30,
		SpeedTestDuration:  time.Second * 20,
		SpeedTestBatchSize: 100,
	}

	survey := survey.NewSurvey(config)
	results := survey.Run()

	fmt.Println("\n\n// Worthy servers:")
	for _, result := range results {
		if result.IsWorthy {
			fmt.Println(toCodeLine(result))
		}
	}

	fmt.Println("\n\n// Unworthy servers:")
	for _, result := range results {
		if !result.IsWorthy {
			fmt.Println(toCodeLine(result))
		}
	}
}

func toCodeLine(r *survey.Result) string {
	if r.Err != nil {
		return fmt.Sprintf("\"%s\", // %v", r.Server, r.Err)
	}

	return fmt.Sprintf(
		"\"%s\", // impl: %s, batching: %v, ttc: %.2f, speed: %d, from: %s",
		r.Server,
		r.Impl,
		r.BatchSupport,
		r.TimeToConnect.Seconds(),
		r.Speed,
		r.FromPeer,
	)
}
