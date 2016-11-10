// Space Rocks Watch is a bot watching
// asteroids coming too close to earth for the incoming week.
// It uses the nasa public API: https://api.nasa.gov/api.html
// You must set the NASA_API_KEY environement variable when
// launching the bot.
// You can get one here: https://api.nasa.gov/index.html#apply-for-an-api-key
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	conf "github.com/dns-gh/flagsconfig"
)

const (
	nasaAsteroidsAPIGet = "https://api.nasa.gov/neo/rest/v1/feed?api_key="
	timeFormat          = "2006-01-02"
	orbitingBodyToWatch = "Earth"
	timeInterval        = 0 // 0 meaning you get the current day info,...
	updateFrequency     = 6 * time.Hour
	fetchMaxSizeError   = "cannot fetch infos for more than 7 days in one request"
)

var (
	envErrorList = []string{}
	nasaAPIKey   = "DEMO_KEY"
	twitterAPI   *anaconda.TwitterApi
)

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		envErrorList = append(envErrorList, fmt.Sprintf("%q is not defined", key))
	}
	return value
}

var (
	twitterConsumerKey    = getEnv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret = getEnv("TWITTER_CONSUMER_SECRET")
	twitterAccessToken    = getEnv("TWITTER_ACCESS_TOKEN")
	twitterAccessSecret   = getEnv("TWITTER_ACCESS_SECRET")
)

func init() {
	key := os.Getenv("NASA_API_KEY")
	if key != "" {
		nasaAPIKey = key
	}
	if len(envErrorList) > 0 {
		log.Fatalln(fmt.Sprintf("errors:\n%s", strings.Join(envErrorList, "\n")))
	}
	anaconda.SetConsumerKey(twitterConsumerKey)
	anaconda.SetConsumerSecret(twitterConsumerSecret)
	twitterAPI = anaconda.NewTwitterApi(twitterAccessToken, twitterAccessSecret)
}

func updateBot(interval int) error {
	err := checkNasaRocks(interval)
	if err != nil {
		return err
	}
	err = checkRetweet()
	if err != nil {
		return err
	}
	return nil
}

func runBot(interval int) error {
	// check one time before launching the ticker
	// since the ticker begins to tick the first
	// time after the given update frequency.
	err := updateBot(interval)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(updateFrequency)
	defer ticker.Stop()
	for _ = range ticker.C {
		err := updateBot(timeInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	interval := flag.Int("offset", timeInterval, "(offset in days) when fetching data for the first time, fetches data in [now, offset] if offset > 0, [offset, now] otherwise")
	_, err := conf.NewConfig("nasa.config")
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("launching nasa space rocks bot...")
	err = runBot(*interval)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
