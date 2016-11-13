package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"net/url"

	"strconv"

	"github.com/ChimeraCoder/anaconda"
	conf "github.com/dns-gh/flagsconfig"
	"github.com/dns-gh/tojson"
)

var (
	searchTweetQueries = []string{
		"nasa mars",
		"nasa simulation",
		"nasa flight",
		"asteroids solar system",
		"asteroids comets",
		"asteroids belt",
		"asteroids ceres",
		"asteroids orbit",
		"asteroids mercury",
		"asteroids venus",
		"asteroids earth",
		"asteroids mars",
		"asteroids jupiter",
		"asteroids saturn",
		"asteroids uranus",
		"asteroids neptune",
		"asteroids around galaxy",
		"asteroids galactic tour",
		"asteroids universe",
		"asteroids space",
		"asteroids nasa",
		"asteroids deadly",
		"asteroids watch danger",
		"asteroids end world",
		"Near-Earth Object Program",
		"asteroids close approach",
		"asteroids strike",
		"asteroids damages on earth",
		"asteroid impact simulation",
		"asteroid impact",
		"asteroids threat",
		"asteroids exploitation",
		"asteroids mining",
		"asteroid discovery",
	}
	asteroidsQualificativeAdjective = []string{
		"harmless",
		"nasty",
		"threatening",
		"dangerous",
		"critical",
		"terrible",
		"bloody",
		"destructive",
		"deadly",
		"fatal",
	}
	maxRetweetBySearch    = 2
	maxFavoriteCountWatch = 2
)

type twitterBot struct {
	twitterClient *anaconda.TwitterApi
	updateFreq    time.Duration
	path          string
	nasaClient    *nasaClient
}

func (t *twitterBot) tweetNasaData(offset int) error {
	diff, err := t.nasaClient.fetch(offset)
	if err != nil {
		return err
	}
	log.Println("tweeting", len(diff), "tweet(s) about rocks...")
	for _, msg := range diff {
		tw := url.Values{}
		tweet, err := t.twitterClient.PostTweet(msg, tw)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Println("tweet: (id:", tweet.Id, "):", trunc(tweet.Text))
	}
	return nil
}

func (t *twitterBot) updateNasa(first bool) {
	log.Println("[nasa] fetching data...")
	offset := t.nasaClient.offset
	if first {
		offset = t.nasaClient.firstOffset
	}
	err := t.tweetNasaData(offset)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("[nasa] fetching data done.")
}

func (t *twitterBot) pollNasa() {
	t.updateNasa(true)
	ticker := time.NewTicker(t.nasaClient.poll)
	defer ticker.Stop()
	for _ = range ticker.C {
		t.updateNasa(false)
	}
}

func (t *twitterBot) update() {
	log.Println("[twitter] updating...")
	err := t.checkRetweet()
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("[twitter] update done.")
}

func (t *twitterBot) run() {
	// polling nasa data
	go func() {
		t.pollNasa()
	}()
	// update twitter
	t.update()
	ticker := time.NewTicker(t.updateFreq)
	defer ticker.Stop()
	for _ = range ticker.C {
		t.update()
	}
}

func getEnv(errorList []string, key string) string {
	value := os.Getenv(key)
	if value == "" {
		errorList = append(errorList, fmt.Sprintf("%q is not defined", key))
	}
	return value
}

func makeTwitterBot(config *conf.Config) *twitterBot {
	errorList := []string{}
	consumerKey := getEnv(errorList, "TWITTER_CONSUMER_KEY")
	consumerSecret := getEnv(errorList, "TWITTER_CONSUMER_SECRET")
	accessToken := getEnv(errorList, "TWITTER_ACCESS_TOKEN")
	accessSecret := getEnv(errorList, "TWITTER_ACCESS_SECRET")
	if len(errorList) > 0 {
		log.Fatalln(fmt.Sprintf("errors:\n%s", strings.Join(errorList, "\n")))
	}
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	return &twitterBot{
		twitterClient: anaconda.NewTwitterApi(accessToken, accessSecret),
		updateFreq:    parseDuration(config.Get(updateFlag)),
		path:          config.Get(twitterPathFlag),
		nasaClient:    makeNasaClient(config),
	}
}

// TODO factorize with load, merge and update ?
func (t *twitterBot) loadTweets(path string) ([]anaconda.Tweet, error) {
	tweets := &[]anaconda.Tweet{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		tojson.Save(path, tweets)
	}
	err := tojson.Load(path, tweets)
	if err != nil {
		return nil, err
	}
	return *tweets, nil
}

func mergeTweets(previous, current []anaconda.Tweet) ([]anaconda.Tweet, []anaconda.Tweet) {
	merged := []anaconda.Tweet{}
	diff := []anaconda.Tweet{}
	added := map[int64]struct{}{}
	for _, v := range previous {
		added[v.Id] = struct{}{}
		merged = append(merged, v)
	}
	for _, v := range current {
		if _, ok := added[v.Id]; ok {
			continue
		}
		added[v.Id] = struct{}{}
		merged = append(merged, v)
		diff = append(diff, v)
	}
	return merged, diff
}

func (t *twitterBot) updateTweets(path string, current []anaconda.Tweet) ([]anaconda.Tweet, error) {
	previous, err := t.loadTweets(path)
	if err != nil {
		return nil, err
	}
	merged, diff := mergeTweets(previous, current)
	tojson.Save(path, merged)
	return diff, nil
}

func (t *twitterBot) getRelevantTweets() ([]anaconda.Tweet, error) {
	query := getRandomElement(searchTweetQueries)
	v := url.Values{}
	v.Set("count", strconv.Itoa(maxRetweetBySearch+2))
	results, _ := t.twitterClient.GetSearch(query, v)
	return results.Statuses, nil
}

func (t *twitterBot) like(tweet *anaconda.Tweet) {
	if tweet.FavoriteCount > maxFavoriteCountWatch {
		_, err := t.twitterClient.Favorite(tweet.Id)
		if err != nil {
			log.Printf("failed to like tweet (id:%d), error: %v\n", tweet.Id, err)
		}
		log.Printf("liked tweet (id:%d): %s\n", tweet.Id, trunc(tweet.Text))
	} else if tweet.RetweetedStatus != nil &&
		tweet.RetweetedStatus.FavoriteCount > maxFavoriteCountWatch {
		t.like(tweet.RetweetedStatus)
	}
}

func (t *twitterBot) checkRetweet() error {
	log.Println("checking tweets to retweet...")
	// TODO some tweet are retweet and hence could be the same on the below list
	current, err := t.getRelevantTweets()
	if err != nil {
		return err
	}
	log.Println("found", len(current), "potential tweet to retweet")
	// TODO only merge and save tweets once they are retweeted ?
	diff, err := t.updateTweets(t.path, current)
	if err != nil {
		return err
	}
	log.Println("retweeting", len(diff), "tweets...")
	for _, tweet := range diff {
		sleep(maxRandTimeSleepBetweenTweets)
		t.like(&tweet)
		retweet, err := t.twitterClient.Retweet(tweet.Id, false)
		if err != nil {
			log.Printf("failed to retweet tweet (id:%d), error: %v\n", tweet.Id, err)
			continue
		}
		t.like(&retweet)
		log.Printf("retweet (r_id:%d, id:%d): %s\n", retweet.Id, tweet.Id, trunc(retweet.Text))
	}
	return nil
}
