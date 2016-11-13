// Space Rocks Watch is a bot watching
// asteroids coming too close to earth for the incoming week.
// It uses the nasa public API: https://api.nasa.gov/api.html
// You must set the NASA_API_KEY environement variable when
// launching the bot.
// You can get one here: https://api.nasa.gov/index.html#apply-for-an-api-key
package main

import (
	"flag"
	"log"
	"time"

	conf "github.com/dns-gh/flagsconfig"
)

const (
	debugFlag = "debug"
)

// Nasa constants
const (
	nasaAsteroidsAPIGet = "https://api.nasa.gov/neo/rest/v1/feed?api_key="
	nasaAPIDefaultKey   = "DEMO_KEY"
	nasaTimeFormat      = "2006-01-02"
	fetchMaxSizeError   = "cannot fetch infos for more than 7 days in one request"
	// flags definitions
	firstOffsetFlag   = "first-offset"
	offsetFlag        = "offset"
	bodyFlag          = "body"
	pollFrequencyFlag = "poll"
	// TODO save only relevant information on asteroids, the file could become too large at some point otherwise
	nasaPathFlag = "nasa-path"
)

// Twitter constants
const (
	updateFlag = "update"
	// TODO save only relevant information on tweets, the file coudl become too large at some point otherwise
	twitterPathFlag  = "twitter-path"
	retweetTextTag   = "RT @"
	retweetTextIndex = ": "
)

var (
	maxRandTimeSleepBetweenTweets = 120 // seconds
)

func main() {
	firstOffset := flag.Int(firstOffsetFlag, 0, "[nasa] offset when fetching data for the first time (days)")
	offset := flag.Int(offsetFlag, 3, "[nasa] offset when fetching data (days)")
	body := flag.String(bodyFlag, "Earth", "[nasa] orbiting body to watch for close asteroids")
	poll := flag.Duration(pollFrequencyFlag, 12*time.Hour, "[nasa] polling frequency of data")
	nasaPath := flag.String(nasaPathFlag, "rocks.json", "[nasa] data file path")
	update := flag.Duration(updateFlag, 2*time.Hour, "[twitter] update frequency of the bot")
	twitterPath := flag.String(twitterPathFlag, "tweets.json", "[twitter] data file path")
	debug := flag.Bool(debugFlag, false, "[twitter] debug mode")
	config, err := conf.NewConfig("nasa.config")
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("[nasa] first-offset:", *firstOffset)
	log.Println("[nasa] offset:", *offset)
	log.Println("[nasa] body:", *body)
	log.Println("[nasa] poll:", *poll)
	log.Println("[nasa] nasa-path:", *nasaPath)
	log.Println("[twitter] update:", *update)
	log.Println("[twitter] twitter-path:", *twitterPath)
	log.Println("[twitter] debug:", *debug)
	bot := makeTwitterBot(config)
	log.Println(" --- launching bot ---")
	bot.run()
}
