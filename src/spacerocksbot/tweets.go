package main

import (
	"log"
	"math/rand"
	"os"

	"net/url"

	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dns-gh/tojson"
)

const (
	tweetsFilePath = "tweets.json"
)

var (
	searchTweetQueries = [...]string{
		"asteroids earth",
		"asteroids mars",
		"asteroids space",
		"asteroids nasa",
		"asteroids deadly",
		"asteroids watch danger",
		"asteroids end world",
	}
	maxRetweetBySearch = 2
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
	query := searchTweetQueries[rand.Intn(len(searchTweetQueries))]
	v := url.Values{}
	v.Set("count", strconv.Itoa(maxRetweetBySearch+2))
	results, _ := twitterAPI.GetSearch(query, v)
	return results.Statuses, nil
}

func checkRetweet() error {
	current, err := getRelevantTweets()
	if err != nil {
		return err
	}
	diff, err := updateTweets(tweetsFilePath, current)
	if err != nil {
		return err
	}
	for _, tweet := range diff {
		sleep(maxRandTimeSleepBetweenTweets)
		retweet, err := twitterAPI.Retweet(tweet.Id, false)
		if err != nil {
			log.Printf("failed to retweet msg for tweet (id:%d), error: %v\n", tweet.Id, err)
			continue
		}
		log.Println("retweet (id:", tweet.Id, "):", retweet.Text)
	}
	return nil
}
