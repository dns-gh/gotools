// Space Rocks Watch is a bot watching
// asteroids coming too close to earth for the incoming week.
// It uses the nasa public API: https://api.nasa.gov/api.html
// You must set the NASA_API_KEY environement variable when
// launching the bot.
// You can get one here: https://api.nasa.gov/index.html#apply-for-an-api-key
package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
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
	updateFlag               = "update"
	twitterFollowersPathFlag = "twitter-followers-path"
	twitterFriendsPathFlag   = "twitter-friends-path"
	// TODO save only relevant information on tweets, the file could become too large at some point otherwise
	twitterTweetsPathFlag       = "twitter-tweets-path"
	maxRetweetBySearch          = 2
	maxFavoriteCountWatch       = 2
	maxTryRetweetCount          = 5
	retweetTextTag              = "RT @"
	retweetTextIndex            = ": "
	tweetHTTPTag                = "http://"
	OneDayInNano          int64 = 86400000000000
)

var (
	maxRandTimeSleepBetweenRequests = 120               // seconds
	timeSleepBetweenFollowUnFollow  = 300 * time.Second // seconds
)

func main() {
	firstOffset := flag.Int(firstOffsetFlag, 0, "[nasa] offset when fetching data for the first time (days)")
	offset := flag.Int(offsetFlag, 3, "[nasa] offset when fetching data (days)")
	body := flag.String(bodyFlag, "Earth", "[nasa] orbiting body to watch for close asteroids")
	poll := flag.Duration(pollFrequencyFlag, 12*time.Hour, "[nasa] polling frequency of data")
	nasaPath := flag.String(nasaPathFlag, "rocks.json", "[nasa] data file path")
	update := flag.Duration(updateFlag, 30*time.Minute, "[twitter] update frequency of the bot")
	twitterFollowersPath := flag.String(twitterFollowersPathFlag, "followers.json", "[twitter] data file path for followers")
	twitterFriendsPath := flag.String(twitterFriendsPathFlag, "friends.json", "[twitter] data file path for friends")
	twitterTweetsPath := flag.String(twitterTweetsPathFlag, "tweets.json", "[twitter] data file path for tweets")
	debug := flag.Bool(debugFlag, false, "[twitter] debug mode")
	config, err := conf.NewConfig("nasa.config")
	// log to a file also
	log.SetFlags(0)
	logPath, f, err := makeLog(filepath.Join(filepath.Dir(os.Args[0]), "Debug", "bot.log"))
	if err == nil {
		defer f.Close()
		log.SetOutput(io.MultiWriter(f, os.Stderr))
	}
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("logging to:", logPath)
	log.Println("[nasa] first-offset:", *firstOffset)
	log.Println("[nasa] offset:", *offset)
	log.Println("[nasa] body:", *body)
	log.Println("[nasa] poll:", *poll)
	log.Println("[nasa] nasa-path:", *nasaPath)
	log.Println("[twitter] update:", *update)
	log.Println("[twitter] twitter-followers-path:", *twitterFollowersPath)
	log.Println("[twitter] twitter-friends-path:", *twitterFriendsPath)
	log.Println("[twitter] twitter-tweets-path:", *twitterTweetsPath)
	log.Println("[twitter] debug:", *debug)

	bot := makeTwitterBot(config, *update, *twitterFollowersPath, *twitterFriendsPath,
		*twitterTweetsPath, searchTweetQueries, *debug)
	defer bot.close()
	log.Println(" --- launching bot ---")
	bot.run()
}
