package main

import (
	"log"
	"os"

	"net/url"

	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dns-gh/tojson"
)

const (
	// TODO save only relevant information on tweets, the file coudl become too large at some point otherwise
	tweetsFilePath = "tweets.json"
)

var (
	searchTweetQueries = []string{
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

// TODO factorize with load, merge and update
func loadTweets(path string) ([]anaconda.Tweet, error) {
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

func updateTweets(path string, current []anaconda.Tweet) ([]anaconda.Tweet, error) {
	previous, err := loadTweets(path)
	if err != nil {
		return nil, err
	}
	merged, diff := mergeTweets(previous, current)
	tojson.Save(path, merged)
	return diff, nil
}

func getRelevantTweets() ([]anaconda.Tweet, error) {
	query := getRandomElement(searchTweetQueries)
	v := url.Values{}
	v.Set("count", strconv.Itoa(maxRetweetBySearch+2))
	results, _ := twitterAPI.GetSearch(query, v)
	return results.Statuses, nil
}

func like(tweet *anaconda.Tweet) {
	if tweet.FavoriteCount > maxFavoriteCountWatch {
		_, err := twitterAPI.Favorite(tweet.Id)
		if err != nil {
			log.Printf("failed to likes/favorites tweet (id:%d), error: %v\n", tweet.Id, err)
		}
	}
}

func checkRetweet() error {
	// TODO some tweet are retweet and hence could be the same on the below list
	current, err := getRelevantTweets()
	if err != nil {
		return err
	}
	// TODO only merge and save tweets once they are retweeted ?
	diff, err := updateTweets(tweetsFilePath, current)
	if err != nil {
		return err
	}
	for _, tweet := range diff {
		sleep(maxRandTimeSleepBetweenTweets)
		like(&tweet)
		retweet, err := twitterAPI.Retweet(tweet.Id, false)
		if err != nil {
			log.Printf("failed to retweet msg for tweet (id:%d), error: %v\n", tweet.Id, err)
			continue
		}
		like(&retweet)
		log.Printf("retweet (r_id:%d, id:%d): %s\n", retweet.Id, tweet.Id, retweet.Text)
	}
	return nil
}
