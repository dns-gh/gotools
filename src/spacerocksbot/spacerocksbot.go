// Space Rocks Watch is a bot watching
// asteroids coming too close to earth for the incoming week.
// It uses the nasa public API: https://api.nasa.gov/api.html
// You must set the NASA_API_KEY environement variable when
// launching the bot.
// You can get one here: https://api.nasa.gov/index.html#apply-for-an-api-key
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	updateFrequency     = 12 * time.Hour
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
func fetchRocks(days int) (*SpaceRocks, error) {
	if days > 7 {
		return nil, fmt.Errorf(fetchMaxSizeError)
	} else if days < -7 {
		return nil, fmt.Errorf(fetchMaxSizeError)
	}
	now := time.Now()
	start := ""
	end := ""
	if days >= 0 {
		start = now.Format(timeFormat)
		end = now.AddDate(0, 0, days).Format(timeFormat)
	} else {
		start = now.AddDate(0, 0, days).Format(timeFormat)
		end = now.Format(timeFormat)
	}
	url := nasaAsteroidsAPIGet +
		nasaAPIKey +
		"&start_date=" + start +
		"&end_date=" + end
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer resp.Body.Close()

	spacerocks := &SpaceRocks{}
	json.NewDecoder(resp.Body).Decode(spacerocks)
	return spacerocks, nil
}

func getDangerousRocks(interval int) ([]object, error) {
	rocks, err := fetchRocks(interval)
	if err != nil {
		return nil, err
	}
	dangerous := map[int64]object{}
	keys := []int64{}
	for _, v := range rocks.NearEarthObjects {
		if len(v) != 0 {
			for _, object := range v {
				if object.IsPotentiallyHazardousAsteroid {
					if len(object.CloseApproachData) != 0 &&
						object.CloseApproachData[0].OrbitingBody == orbitingBodyToWatch {
						t, err := parseTime(object.CloseApproachData[0].CloseApproachDate)
						if err != nil {
							return nil, err
						}
						timestamp := t.UnixNano()
						dangerous[timestamp] = object
						keys = append(keys, timestamp)
					}
				}
			}
		}
	}
	quickSort(keys)
	objects := []object{}
	for _, key := range keys {
		objects = append(objects, dangerous[key])
	}
	return objects, nil
}

func checkNasaRocks(interval int) error {
	current, err := getDangerousRocks(interval)
	if err != nil {
		return err
	}
	diff, err := update(current)
	if err != nil {
		return err
	}
	fmt.Println("+--------------------------------------------------------------+")
	fmt.Println("|               Potential dangerous incoming rocks             |")
	fmt.Println("+--------------------------------------------------------------+")
	for _, object := range diff {
		t, err := parseTime(object.CloseApproachData[0].CloseApproachDate)
		if err != nil {
			return err
		}
		statusMsg := fmt.Sprintf("a dangerous asteroid of Ø %.2f to %.2f km is coming near %s on %d-%02d-%02d \n",
			object.EstimatedDiameter.Kilometers.EstimatedDiameterMin,
			object.EstimatedDiameter.Kilometers.EstimatedDiameterMax,
			orbitingBodyToWatch,
			t.Year(), t.Month(), t.Day())
		tw := url.Values{}
		tweet, err := twitterAPI.PostTweet(statusMsg, tw)
		if err != nil {
			log.Println("failed to tweet msg for object id:", object.NeoReferenceID)
		}
		log.Println("tweet:", tweet.Text)
	}
	fmt.Println("+--------------------------------------------------------------+")
	return nil
}

func runBot(interval int) error {
	// check one time before launching the ticker
	// since the ticker begins to tick the first
	// time after the given update frequency.
	err := checkNasaRocks(interval)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(updateFrequency)
	defer ticker.Stop()
	for _ = range ticker.C {
		err := checkNasaRocks(timeInterval)
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
	err = runBot(*interval)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
