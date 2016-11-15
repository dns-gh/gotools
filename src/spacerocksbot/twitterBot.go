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
)

type twitterBot struct {
	twitterClient *anaconda.TwitterApi
	updateFreq    time.Duration
	path          string
	nasaClient    *nasaClient
	debug         bool
}

func (t *twitterBot) tweetNasaData(offset int) error {
	diff, err := t.nasaClient.fetch(offset)
	if err != nil {
		return err
	}
	log.Println("[twitter] tweeting", len(diff), "tweet(s) about rocks...")
	for _, msg := range diff {
		tweet, err := t.twitterClient.PostTweet(msg, nil)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Println("tweet: (id:", tweet.Id, "):", trunc(tweet.Text))
		t.like(&tweet)
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
	log.Println("[twitter] updating... done.")
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
	bot := &twitterBot{
		twitterClient: anaconda.NewTwitterApi(accessToken, accessSecret),
		updateFreq:    parseDuration(config.Get(updateFlag)),
		path:          config.Get(twitterPathFlag),
		nasaClient:    makeNasaClient(config),
		debug:         parseBool(config.Get(debugFlag)),
	}
	go func() {
		//bot.unfollowAll()
		log.Println("[twitter] auto unfollow disabled")
	}()
	go func() {
		//time.Sleep(45 * time.Second)
		//bot.followAll()
		log.Println("[twitter] auto follow disabled")
	}()
	return bot
}

func (t *twitterBot) close() {
	t.twitterClient.Close()
}

// TODO factorize with load, merge and update ?
func (t *twitterBot) loadTweets() ([]anaconda.Tweet, error) {
	tweets := &[]anaconda.Tweet{}
	if _, err := os.Stat(t.path); os.IsNotExist(err) {
		tojson.Save(t.path, tweets)
	}
	err := tojson.Load(t.path, tweets)
	if err != nil {
		return nil, err
	}
	return *tweets, nil
}

func (t *twitterBot) getOriginalText(tweet *anaconda.Tweet) string {
	text := tweet.Text
	if strings.Contains(tweet.Text, retweetTextTag) {
		tab := strings.SplitN(text, retweetTextIndex, 2)
		if len(tab) != 2 {
			log.Println("[twitter] error parsing a tweet text:", text)
			return text
			// TODO do something
		}
		text = tab[1]
		if strings.Contains(text, tweetHTTPTag) {
			subtab := strings.SplitN(text, tweetHTTPTag, 2)
			if len(subtab) > 2 {
				log.Println("[twitter] error parsing a sub tweet text:", text)
				return text
				// TODO do something
			}
			text = subtab[0]
		}
	}
	return text
}

func (t *twitterBot) takeDifference(previous, current []anaconda.Tweet) []anaconda.Tweet {
	diff := []anaconda.Tweet{}
	addedByID := map[int64]struct{}{}
	addedByText := map[string]struct{}{}
	for _, v := range previous {
		addedByID[v.Id] = struct{}{}
		addedByText[t.getOriginalText(&v)] = struct{}{}
	}
	for _, v := range current {
		if _, ok := addedByID[v.Id]; ok {
			log.Printf("[twitter] found a duplicate (same id) from database id:%d, text:%s", v.Id, v.Text)
			continue
		}
		text := t.getOriginalText(&v)
		if _, ok := addedByText[text]; ok {
			log.Printf("[twitter] found a duplicate (same original text) from database id:%d, text:%s", v.Id, v.Text)
			continue
		}
		addedByID[v.Id] = struct{}{}
		addedByText[text] = struct{}{}
		diff = append(diff, v)
	}
	return diff
}

func (t *twitterBot) removeDuplicates(current []anaconda.Tweet) []anaconda.Tweet {
	temp := map[string]struct{}{}
	stripped := []anaconda.Tweet{}
	for _, tweet := range current {
		text := t.getOriginalText(&tweet)
		if _, ok := temp[text]; !ok {
			temp[text] = struct{}{}
			stripped = append(stripped, tweet)
		} else {
			log.Printf("[twitter] found a duplicate (id:%d), text:%s", tweet.Id, tweet.Text)
		}
	}
	return stripped
}

func (t *twitterBot) getRelevantTweets() ([]anaconda.Tweet, error) {
	query := getRandomElement(searchTweetQueries)
	log.Println("searching with query:", query)
	v := url.Values{}
	v.Set("count", strconv.Itoa(maxRetweetBySearch+2))
	results, _ := t.twitterClient.GetSearch(query, v)
	return results.Statuses, nil
}

func (t *twitterBot) like(tweet *anaconda.Tweet) {
	if tweet.FavoriteCount > maxFavoriteCountWatch {
		_, err := t.twitterClient.Favorite(tweet.Id)
		if err != nil {
			log.Printf("[twitter] failed to like tweet (id:%d), error: %v\n", tweet.Id, err)
		}
		log.Printf("[twitter] liked tweet (id:%d): %s\n", tweet.Id, trunc(tweet.Text))
	} else if tweet.RetweetedStatus != nil &&
		tweet.RetweetedStatus.FavoriteCount > maxFavoriteCountWatch {
		t.like(tweet.RetweetedStatus)
	}
}

func (t *twitterBot) sleep() {
	if !t.debug {
		sleep(maxRandTimeSleepBetweenTweets)
	}
}

func (t *twitterBot) unfollowUser(user *anaconda.User) {
	unfollowed, err := t.twitterClient.UnfollowUserId(user.Id)
	if err != nil {
		log.Printf("[twitter] failed to unfollow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err)
	}
	log.Printf("[twitter] unfollowing user (id:%d, name:%s)\n", unfollowed.Id, unfollowed.Name)
}

func (t *twitterBot) followUser(user *anaconda.User) {
	followed, err := t.twitterClient.FollowUserId(user.Id, nil)
	if err != nil {
		log.Printf("[twitter] failed to follow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err)
	}
	log.Printf("[twitter] following user (id:%d, name:%s)\n", followed.Id, followed.Name)
	//t.unfollowUser(&followed)
}

func (t *twitterBot) retweet(current []anaconda.Tweet) (rt anaconda.Tweet, err error) {
	for _, tweet := range current {
		log.Printf("[twitter] trying to retweet tweet id:%d\n", tweet.Id)
		//t.sleep()
		t.like(&tweet)
		retweet, err := t.twitterClient.Retweet(tweet.Id, false)
		if err != nil {
			log.Printf("[twitter] failed to retweet tweet (id:%d), error: %v\n", tweet.Id, err)
			t.followUser(&tweet.User)
			continue
		}
		rt = retweet
		t.like(&retweet)
		log.Printf("[twitter] retweet (r_id:%d, id:%d): %s\n", retweet.Id, tweet.Id, trunc(retweet.Text))
		t.followUser(&tweet.User)
		return rt, err
	}
	err = fmt.Errorf("unable to retweet")
	return rt, err
}

func (t *twitterBot) getTweets(previous []anaconda.Tweet) ([]anaconda.Tweet, error) {
	log.Println("[twitter] checking tweets to retweet...")
	current, err := t.getRelevantTweets()
	if err != nil {
		return nil, err
	}
	current = t.removeDuplicates(current)
	current = t.takeDifference(previous, current)
	log.Println("[twitter] found", len(current), "tweet(s) matching pattern")
	return current, nil
}

func (t *twitterBot) checkRetweet() error {
	count := 0
	previous, err := t.loadTweets()
	if err != nil {
		return err
	}
	for {
		tweets, err := t.getTweets(previous)
		if err != nil {
			return err
		}
		retweeted, err := t.retweet(tweets)
		if err != nil {
			if count < maxTryRetweetCount {
				count++
				continue
			} else {
				return fmt.Errorf("[twitter] unable to retweet something after %d tries\n", maxTryRetweetCount)
			}
		}
		previous = append(previous, retweeted)
		tojson.Save(t.path, previous)
		return nil
	}
}

func (t *twitterBot) unfollowAll() {
	time.Sleep(90 * time.Second)
	cursor, err := t.twitterClient.GetFriendsIds(nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, id := range cursor.Ids {
		time.Sleep(50 * time.Second)
		user, err := t.twitterClient.UnfollowUserId(id)
		if err != nil {
			log.Println(err.Error())
			if strings.Contains(err.Error(), "Invalid or expired token") {
				return
			}
		}
		log.Printf("[twitter] unfollowing (id:%d, name:%s)\n", user.Id, user.Name)
	}
	t.unfollowAll()
}

func (t *twitterBot) followAll() {
	time.Sleep(90 * time.Second)
	ids := t.getFollowersIds("nasa", 0)
	t.followIds(ids)
	t.followAll()
}

func (t *twitterBot) followIds(ids map[int64]struct{}) {
	for id := range ids {
		time.Sleep(90 * time.Second)
		user, err := t.twitterClient.FollowUserId(id, nil)
		if err != nil {
			log.Printf("[twitter] failed to follow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err)
			if strings.Contains(err.Error(), "You are unable to follow more people at this time") {
				break
			}
			continue
		}
		log.Printf("[twitter] following (id:%d, name:%s)\n", user.Id, user.Name)
	}
}

func (t *twitterBot) getFollowersIds(query string, depth int) map[int64]struct{} {
	fmt.Printf("[twitter] searching people to follow (q:%s, depth:%d)\n", query, depth)
	users, err := t.twitterClient.GetUserSearch(query, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	ids := map[int64]struct{}{}
	count := 0
	for _, user := range users {
		nextCursor := "-1"
		currentDepth := 0
		for {
			v := url.Values{}
			if nextCursor != "-1" {
				v.Set("cursor", nextCursor)
			}
			cursor, err := t.twitterClient.GetFollowersUser(user.Id, nil)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			for _, v := range cursor.Ids {
				ids[v] = struct{}{}
			}
			if currentDepth >= depth {
				break
			}
			currentDepth++
			nextCursor = cursor.Next_cursor_str
			if nextCursor == "0" {
				break
			}
		}
		count++
		if count >= 1 /*getUserSearchAPIRateLimit*/ {
			break
		}
	}
	return ids
}
