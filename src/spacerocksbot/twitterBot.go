package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"net/url"

	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dns-gh/tojson"
)

type twitterUser struct {
	Timestamp int64 `json:"timestamp"`
	Follow    bool  `json:"follow"`
}

type twitterUsers struct {
	// note: we cannot use integers as keys in encode/json so use string instead
	Ids map[string]*twitterUser `json:"ids"` // map id -> user
}

type likePolicy struct {
	auto      bool
	threshold int
}

type twitterBot struct {
	twitterClient *anaconda.TwitterApi
	followersPath string
	followers     *twitterUsers
	friendsPath   string
	friends       *twitterUsers
	tweetsPath    string
	debug         bool
	likePolicy    *likePolicy
	mutex         sync.Mutex
	quit          sync.WaitGroup
}

func makeTwitterBot(followersPath, friendsPath, tweetsPath string, autoLike bool, autoThreshold int, debug bool) *twitterBot {
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
		followersPath: followersPath,
		followers: &twitterUsers{
			Ids: make(map[string]*twitterUser),
		},
		friendsPath: friendsPath,
		friends: &twitterUsers{
			Ids: make(map[string]*twitterUser),
		},
		tweetsPath: tweetsPath,
		debug:      debug,
		likePolicy: &likePolicy{
			auto:      autoLike,
			threshold: autoThreshold,
		},
	}
	err := bot.updateFollowers()
	if err != nil {
		log.Println(err.Error())
	}
	err = bot.updateFriends()
	if err != nil {
		log.Println(err.Error())
	}
	go func() {
		log.Println("[twitter] launching auto unfollow...")
		bot.unfollowAll()
		log.Println("[twitter] - WARNING - auto unfollow disabled")
	}()
	go func() {
		log.Println("[twitter] launching auto follow...")
		ids := bot.fetchUserIds("nasa", 0)
		bot.followAll(ids)
		log.Println("[twitter] - WARNING - auto follow disabled")
	}()
	return bot
}

func (t *twitterBot) Wait() {
	t.quit.Wait()
}

func (t *twitterBot) Close() {
	t.twitterClient.Close()
}

func (t *twitterBot) isFollower(id int64) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	_, ok := t.followers.Ids[strconv.FormatInt(id, 10)]
	return ok
}

func (t *twitterBot) getFriend(id int64) (*twitterUser, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	user, ok := t.friends.Ids[strconv.FormatInt(id, 10)]
	if ok {
		return &twitterUser{
			Timestamp: user.Timestamp,
			Follow:    user.Follow,
		}, ok
	}
	return nil, false
}

func (t *twitterBot) getFriendToUnFollow() (int64, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for strID, user := range t.friends.Ids {
		// unfollow only if is followed and is in database from at least 1 day
		if time.Now().UnixNano()-user.Timestamp < OneDayInNano || !user.Follow {
			continue
		}
		id, err := strconv.ParseInt(strID, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		return id, true
	}
	return 0, false
}

func (t *twitterBot) addFriend(id int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.friends.Ids[strconv.FormatInt(id, 10)] = &twitterUser{
		Timestamp: time.Now().UnixNano(),
		Follow:    true,
	}
	err := tojson.Save(t.friendsPath, t.friends)
	if err != nil {
		log.Fatalln(err)
	}
}

// UnfollowFriend flags the friend as not followed anymore.
// We do not remove friends from database.
func (t *twitterBot) unfollowFriend(id int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.friends.Ids[strconv.FormatInt(id, 10)].Follow = false
	err := tojson.Save(t.friendsPath, t.friends)
	if err != nil {
		log.Fatalln(err)
	}
}

func (t *twitterBot) tweetSliceOnce(fetch func() ([]string, error)) {
	list, err := fetch()
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, msg := range list {
		tweet, err := t.twitterClient.PostTweet(msg, nil)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Println("tweeting message (id:", tweet.Id, "):", trunc(tweet.Text))
	}
}

func (t *twitterBot) TweetSliceOnce(fetch func() ([]string, error)) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		list, err := fetch()
		if err != nil {
			log.Println(err.Error())
			return
		}
		for _, msg := range list {
			tweet, err := t.twitterClient.PostTweet(msg, nil)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			log.Println("tweeting message (id:", tweet.Id, "):", trunc(tweet.Text))
		}
	}()
}

func (t *twitterBot) TweetSlicePeriodically(fetch func() ([]string, error), freq time.Duration) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		ticker := time.NewTicker(freq)
		defer ticker.Stop()
		for _ = range ticker.C {
			t.tweetSliceOnce(fetch)
		}
	}()
}

func (t *twitterBot) tweetOnce(fetch func() (string, error)) {
	msg, err := fetch()
	if err != nil {
		log.Println(err.Error())
		return
	}
	tweet, err := t.twitterClient.PostTweet(msg, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("tweeting message (id:", tweet.Id, "):", trunc(tweet.Text))
}

func (t *twitterBot) TweetOnce(fetch func() (string, error)) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		t.tweetOnce(fetch)
	}()
}

func (t *twitterBot) TweetPeriodically(fetch func() (string, error), freq time.Duration) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		ticker := time.NewTicker(freq)
		defer ticker.Stop()
		for _ = range ticker.C {
			t.tweetOnce(fetch)
		}
	}()
}

func (t *twitterBot) retweetOnce(queries []string) {
	err := t.autoRetweet(queries)
	if err != nil {
		log.Println(err.Error())
	}
}

func (t *twitterBot) RetweetOnce(searchQueries []string) {
	queries := make([]string, len(searchQueries))
	copy(queries, searchQueries)
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		t.retweetOnce(queries)
	}()
}

func (t *twitterBot) RetweetPeriodically(searchQueries []string, freq time.Duration) {
	queries := make([]string, len(searchQueries))
	copy(queries, searchQueries)
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		ticker := time.NewTicker(freq)
		defer ticker.Stop()
		for _ = range ticker.C {
			t.retweetOnce(queries)
		}
	}()
}

func getEnv(errorList []string, key string) string {
	value := os.Getenv(key)
	if value == "" {
		errorList = append(errorList, fmt.Sprintf("%q is not defined", key))
	}
	return value
}

func (t *twitterBot) loadTweets() ([]anaconda.Tweet, error) {
	tweets := &[]anaconda.Tweet{}
	if _, err := os.Stat(t.tweetsPath); os.IsNotExist(err) {
		tojson.Save(t.tweetsPath, tweets)
	}
	err := tojson.Load(t.tweetsPath, tweets)
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
		}
		text = tab[1]
		if strings.Contains(text, tweetHTTPTag) {
			subtab := strings.SplitN(text, tweetHTTPTag, 2)
			if len(subtab) > 2 {
				log.Println("[twitter] error parsing a sub tweet text:", text)
				return text
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

func (t *twitterBot) like(tweet *anaconda.Tweet) {
	if !t.likePolicy.auto {
		return
	}
	if tweet.FavoriteCount > t.likePolicy.threshold {
		_, err := t.twitterClient.Favorite(tweet.Id)
		if err != nil {
			log.Printf("[twitter] failed to like tweet (id:%d), error: %v\n", tweet.Id, err)
		}
		log.Printf("[twitter] liked tweet (id:%d): %s\n", tweet.Id, trunc(tweet.Text))
	} else if tweet.RetweetedStatus != nil &&
		tweet.RetweetedStatus.FavoriteCount > t.likePolicy.threshold {
		t.like(tweet.RetweetedStatus)
	}
}

func (t *twitterBot) sleep() {
	if !t.debug {
		sleep(maxRandTimeSleepBetweenRequests)
	}
}

func (t *twitterBot) maybeSleep(chance, totalChance, min, max int) {
	if !t.debug {
		maybeSleepMinMax(chance, totalChance, min, max)
	}
}

func (t *twitterBot) unfollowUser(user *anaconda.User) {
	unfollowed, err := t.twitterClient.UnfollowUserId(user.Id)
	if err != nil {
		checkBotRestriction(err)
		log.Printf("[twitter] failed to unfollow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err)
	}
	log.Printf("[twitter] unfollowing user (id:%d, name:%s)\n", unfollowed.Id, unfollowed.Name)
}

func (t *twitterBot) followUser(user *anaconda.User) {
	followed, err := t.twitterClient.FollowUserId(user.Id, nil)
	if err != nil && !checkUnableToFollowAtThisTime(err) {
		checkBotRestriction(err)
		log.Printf("[twitter] failed to follow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err)
	}
	log.Printf("[twitter] following user (id:%d, name:%s)\n", followed.Id, followed.Name)
}

func (t *twitterBot) retweet(current []anaconda.Tweet) (rt anaconda.Tweet, err error) {
	for _, tweet := range current {
		log.Printf("[twitter] trying to retweet tweet id:%d\n", tweet.Id)
		t.like(&tweet)
		retweet, err := t.twitterClient.Retweet(tweet.Id, false)
		if err != nil {
			log.Printf("[twitter] failed to retweet tweet (id:%d), error: %v\n", tweet.Id, err)
			t.followUser(&tweet.User)
			continue
		}
		rt = retweet
		t.like(&rt)
		log.Printf("[twitter] retweet (r_id:%d, id:%d): %s\n", rt.Id, tweet.Id, trunc(rt.Text))
		t.followUser(&tweet.User)
		return rt, err
	}
	err = fmt.Errorf("unable to retweet")
	return rt, err
}

func (t *twitterBot) getTweets(queries []string, previous []anaconda.Tweet) ([]anaconda.Tweet, error) {
	log.Println("[twitter] checking tweets to retweet...")
	query := getRandomElement(queries)
	log.Println("[twitter] searching with query:", query)
	v := url.Values{}
	v.Set("count", strconv.Itoa(maxRetweetBySearch+2))
	results, err := t.twitterClient.GetSearch(query, v)
	if err != nil {
		return nil, err
	}
	current := results.Statuses
	current = t.removeDuplicates(current)
	current = t.takeDifference(previous, current)
	log.Println("[twitter] found", len(current), "tweet(s) matching pattern")
	return current, nil
}

func (t *twitterBot) autoRetweet(queries []string) error {
	count := 0
	previous, err := t.loadTweets()
	if err != nil {
		return err
	}
	for {
		t.sleep()
		tweets, err := t.getTweets(queries, previous)
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
		tojson.Save(t.tweetsPath, previous)
		return nil
	}
}

func (t *twitterBot) updateFollowers() error {
	followers := &twitterUsers{
		Ids: make(map[string]*twitterUser),
	}
	if _, err := os.Stat(t.followersPath); os.IsNotExist(err) {
		tojson.Save(t.followersPath, followers)
	}
	err := tojson.Load(t.followersPath, followers)
	if err != nil {
		return err
	}
	for _, v := range followers.Ids {
		v.Follow = false
	}
	for v := range t.twitterClient.GetFollowersIdsAll(nil) {
		for _, id := range v.Ids {
			strID := strconv.FormatInt(id, 10)
			user, ok := followers.Ids[strID]
			if ok {
				user.Follow = true
			} else {
				followers.Ids[strID] = &twitterUser{
					Timestamp: time.Now().UnixNano(),
					Follow:    true,
				}
			}
		}
	}
	err = tojson.Save(t.followersPath, followers)
	if err != nil {
		return err
	}
	t.followers = followers
	return nil
}

func (t *twitterBot) updateFriends() error {
	friends := &twitterUsers{
		Ids: make(map[string]*twitterUser),
	}
	if _, err := os.Stat(t.friendsPath); os.IsNotExist(err) {
		tojson.Save(t.friendsPath, friends)
	}
	err := tojson.Load(t.friendsPath, friends)
	if err != nil {
		return err
	}
	for _, v := range friends.Ids {
		v.Follow = false
	}
	for v := range t.twitterClient.GetFriendsIdsAll(nil) {
		for _, id := range v.Ids {
			strID := strconv.FormatInt(id, 10)
			user, ok := friends.Ids[strID]
			if ok {
				user.Follow = true
			} else {
				friends.Ids[strID] = &twitterUser{
					Timestamp: time.Now().UnixNano(),
					Follow:    true,
				}
			}
		}
	}
	err = tojson.Save(t.friendsPath, friends)
	if err != nil {
		return err
	}
	t.friends = friends
	return nil
}

func checkBotRestriction(err error) {
	if err != nil {
		strErr := err.Error()
		if strings.Contains(strErr, "Invalid or expired token") ||
			strings.Contains(strErr, "this account is temporarily locked") {
			log.Fatalln(err)
		}
		log.Println(strErr)
	}
}

func checkUnableToFollowAtThisTime(err error) bool {
	if err != nil {
		if strings.Contains(err.Error(), "You are unable to follow more people at this time") {
			log.Println("unable to follow at this time, waiting 15min...,", err.Error())
			time.Sleep(15 * time.Minute)
			return true
		}
		return false
	}
	return false
}

func (t *twitterBot) unfollowAll() {
	var id int64
	for ok := true; ok; id, ok = t.getFriendToUnFollow() {
		if !ok {
			break
		}
		user, err := t.twitterClient.UnfollowUserId(id)
		if err != nil {
			checkBotRestriction(err)
			continue
		}
		t.unfollowFriend(id)
		log.Printf("[twitter] unfollowing (id:%d, name:%s)\n", user.Id, user.Name)
		time.Sleep(timeSleepBetweenFollowUnFollow)
		t.sleep()
		t.maybeSleep(1, 10, 2500, 5000)
	}
	log.Println("[twitter] no more friends to unfollow, waiting 6 hours...")
	time.Sleep(6 * time.Hour)
	t.unfollowAll()
}

func (t *twitterBot) followAll(ids []int64) {
	for _, id := range ids {
		if _, ok := t.getFriend(id); ok || t.isFollower(id) {
			continue
		}
		user, err := t.twitterClient.FollowUserId(id, nil)
		if err != nil && !checkUnableToFollowAtThisTime(err) {
			checkBotRestriction(err)
			log.Printf("[twitter] failed to follow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err)
			continue
		}
		t.addFriend(id)
		log.Printf("[twitter] following (id:%d, name:%s)\n", user.Id, user.Name)
		time.Sleep(timeSleepBetweenFollowUnFollow)
		t.sleep()
		t.maybeSleep(1, 5, 5000, 10000)
	}
}

func (t *twitterBot) fetchUserIds(query string, maxPage int) []int64 {
	fmt.Printf("[twitter] searching people to follow (q:%s, depth:%d)\n", query, maxPage)
	users, err := t.twitterClient.GetUserSearch(query, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	ids := []int64{}
	if len(users) == 0 {
		return nil
	}
	// gettings followers of the first user to avoid a too large volume of users
	user := users[0]
	nextCursor := "-1"
	currentPage := 0
	for {
		v := url.Values{}
		if nextCursor != "-1" {
			v.Set("cursor", nextCursor)
		}
		cursor, err := t.twitterClient.GetFollowersUser(user.Id, nil)
		if err != nil {
			checkBotRestriction(err)
			continue
		}
		for _, v := range cursor.Ids {
			ids = append(ids, v)
		}
		if currentPage >= maxPage {
			break
		}
		currentPage++
		nextCursor = cursor.Next_cursor_str
		if nextCursor == "0" {
			break
		}
	}
	return ids
}
