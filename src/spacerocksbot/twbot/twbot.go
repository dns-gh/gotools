package twbot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"net/url"

	"spacerocksbot/helpers"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dns-gh/tojson"
)

const (
	defaultAutoLikeThreshold              = 1000
	defaultMaxRetweetBySearch             = 5 // keep 3 tweets, the 2 first tweets being useless ?
	retweetTextTag                        = "RT @"
	retweetTextIndex                      = ": "
	tweetHTTPTag                          = "http://"
	oneDayInNano                    int64 = 86400000000000
	timeSleepBetweenFollowUnFollow        = 300 * time.Second // seconds
	maxRandTimeSleepBetweenRequests       = 120               // seconds
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

type retweetPolicy struct {
	maxTry int
	like   bool
}

// TwitterBot represents the twitter bot:
// * The database is made of 3 files: followers, friends and tweets.
// * They are here to ensure to:
//   - not add as friend a friend
//   - not remove as friend a non friend
//   - not retweet a tweet already retweeted
//   - keep track of friends to remove them properly at a specific time if wanted
// * The like policy allows to automatically likes tweets that are already liked above a threshold.
// * The retweet policy allows to try to retweet 'maxTry' times when looping through
//   a list of tweets to retweet. The 'likeRetweet' parameter controls the ability to like the tweet
//   or the retweet using the like policy.
// * The 'debug' mode creates more logs and remove all sleeps between API twitter calls.
type TwitterBot struct {
	twitterClient *anaconda.TwitterApi
	followersPath string
	followers     *twitterUsers
	friendsPath   string
	friends       *twitterUsers
	tweetsPath    string
	debug         bool
	likePolicy    *likePolicy
	retweetPolicy *retweetPolicy
	mutex         sync.Mutex
	quit          sync.WaitGroup
}

// MakeTwitterBot creates a twitter bot
// You have to set up 4 environement variable:
// - TWITTER_CONSUMER_KEY
// - TWITTER_CONSUMER_SECRET
// - TWITTER_ACCESS_TOKEN
// - TWITTER_ACCESS_SECRET
// They can be found here by creating a twitter app: https://apps.twitter.com/
func MakeTwitterBot(followersPath, friendsPath, tweetsPath string, debug bool) *TwitterBot {
	log.Println("[twitter] making twitter bot")
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
	bot := &TwitterBot{
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
			auto:      false,
			threshold: 1000,
		},
		retweetPolicy: &retweetPolicy{
			maxTry: 5,
			like:   true,
		},
	}
	err := bot.updateFollowers()
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = bot.updateFriends()
	if err != nil {
		log.Fatalln(err.Error())
	}
	return bot
}

// Wait waits for all the asynchrones launched tasks to return
func (t *TwitterBot) Wait() {
	t.quit.Wait()
}

// Close closes the twitter client
func (t *TwitterBot) Close() {
	t.twitterClient.Close()
}

// SetLikePolicy sets the like policy
func (t *TwitterBot) SetLikePolicy(auto bool, threshold int) {
	log.Printf("[twitter] setting like policy -> auto: %t, threshold: %d\n", auto, threshold)
	t.likePolicy.auto = auto
	t.likePolicy.threshold = threshold
}

// SetRetweetPolicy sets the retweet policy
func (t *TwitterBot) SetRetweetPolicy(maxTry int, like bool) {
	log.Printf("[twitter] setting retweet policy -> maxTry: %d, like: %t\n", maxTry, like)
	t.retweetPolicy.maxTry = maxTry
	t.retweetPolicy.like = like
}

// TweetSliceOnce tweets the slice returned by the given 'fetch' callback.
// It returns an error is the 'fetch' calls fails and only logs errors
// for each failed tweet tentative.
func (t *TwitterBot) TweetSliceOnce(fetch func() ([]string, error)) error {
	list, err := fetch()
	if err != nil {
		return err
	}
	for _, msg := range list {
		tweet, err := t.twitterClient.PostTweet(msg, nil)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		log.Println("tweeting message (id:", tweet.Id, "):", tweet.Text)
	}
	return nil
}

// TweetSliceOnceAsync tweets asynchronously the slice returned by the
// given 'fetch' callback.
// It logs errors for each failed tweet tentative.
func (t *TwitterBot) TweetSliceOnceAsync(fetch func() ([]string, error)) {
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
			t.print(fmt.Sprintf("tweeting message (id: %d): %s\n", tweet.Id, tweet.Text))
		}
	}()
}

// TweetSlicePeriodicallyAsync tweets asynchronously and periodically the
// slice returned by the given 'fetch' callback.
// The slice tweet frequencies is set up by the given 'freq' input parameter.
// It logs errors for each failed tweet tentative.
func (t *TwitterBot) TweetSlicePeriodicallyAsync(fetch func() ([]string, error), freq time.Duration) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		ticker := time.NewTicker(freq)
		defer ticker.Stop()
		for _ = range ticker.C {
			err := t.TweetSliceOnce(fetch)
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

// TweetOnce tweets the message returned by the 'fetch' callback.
// It returns an error if the 'fetch' call failed or if the tweet
// itself failed.
func (t *TwitterBot) TweetOnce(fetch func() (string, error)) error {
	msg, err := fetch()
	if err != nil {
		return err
	}
	tweet, err := t.twitterClient.PostTweet(msg, nil)
	if err != nil {
		return err
	}
	t.print(fmt.Sprintf("tweeting message (id: %d): %s\n", tweet.Id, tweet.Text))
	return nil
}

// TweetOnceAsync tweets asynchronously the message returned by the 'fetch' callback.
// It only logs the error if the 'fetch' call failed or if the tweet itself failed.
func (t *TwitterBot) TweetOnceAsync(fetch func() (string, error)) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		err := t.TweetOnce(fetch)
		if err != nil {
			log.Println(err)
		}
	}()
}

// TweetPeriodicallyAsync tweets asynchronously and periodically the message returned
// by the 'fetch' callback.
// The tweet frequencies is set up by the given 'freq' input parameter.
// It only logs the error if the 'fetch' call failed or if the tweet itself failed.
func (t *TwitterBot) TweetPeriodicallyAsync(fetch func() (string, error), freq time.Duration) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		ticker := time.NewTicker(freq)
		defer ticker.Stop()
		for _ = range ticker.C {
			err := t.TweetOnce(fetch)
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

// RetweetOnce retweets randomly, with a maximum of 'retweetPolicy.maxTry' tries,
// a tweet matching one element of the input queries slice.
// It returns an error if the loading of tweets in database failed
// or if the retweet itself failed.
func (t *TwitterBot) RetweetOnce(queries []string) error {
	err := t.autoRetweet(queries)
	if err != nil {
		return err
	}
	return nil
}

// RetweetOnceAsync retweets asynchronously and randomly, with a maximum of
// 'retweetPolicy.maxTry' tries, a tweet matching one element of the input queries slice.
// It logs errors if the loading of tweets in database failed
// or if the retweets itself failed.
func (t *TwitterBot) RetweetOnceAsync(searchQueries []string) {
	queries := make([]string, len(searchQueries))
	copy(queries, searchQueries)
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		err := t.RetweetOnce(queries)
		if err != nil {
			log.Println(err)
		}
	}()
}

// RetweetPeriodicallyAsync retweets asynchronously, periodically and randomly, with a maximum of
// 'retweetPolicy.maxTry' tries, a tweet matching one element of the input queries slice.
// The retweet frequencies is set up by the given 'freq' input parameter.
// It logs errors if the loading of tweets in database failed
// or if the retweets itself failed.
func (t *TwitterBot) RetweetPeriodicallyAsync(searchQueries []string, freq time.Duration) {
	queries := make([]string, len(searchQueries))
	copy(queries, searchQueries)
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		ticker := time.NewTicker(freq)
		defer ticker.Stop()
		for _ = range ticker.C {
			err := t.RetweetOnce(queries)
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

// AutoUnfollowFriendsAsync automatically asynchronously unfollows friends
// from database that were added at least a day ago by default.
func (t *TwitterBot) AutoUnfollowFriendsAsync() {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		log.Println("[twitter] launching auto unfollow...")
		t.unfollowAll()
		log.Println("[twitter] auto unfollow disabled")
	}()
}

// AutoFollowFollowersAsync automatically asynchronously follow the
// followers of the first user fecthed using the given 'query'.
// The 'maxPage' parameter indicates the number of page of followers
// (5000 users max by page) we want to fetch.
func (t *TwitterBot) AutoFollowFollowersAsync(query string, maxPage int) {
	t.quit.Add(1)
	go func() {
		defer t.quit.Done()
		log.Printf("[twitter] launching auto follow with '%s' over %d page(s)...\n", query, maxPage)
		t.followAll(t.fetchUserIds(query, maxPage))
		log.Println("[twitter] auto follow disabled")
	}()
}

func getEnv(errorList []string, key string) string {
	value := os.Getenv(key)
	if value == "" {
		errorList = append(errorList, fmt.Sprintf("%q is not defined", key))
	}
	return value
}

func (t *TwitterBot) loadTweets() ([]anaconda.Tweet, error) {
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

func (t *TwitterBot) getOriginalText(tweet *anaconda.Tweet) string {
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

func (t *TwitterBot) takeDifference(previous, current []anaconda.Tweet) []anaconda.Tweet {
	diff := []anaconda.Tweet{}
	addedByID := map[int64]struct{}{}
	addedByText := map[string]struct{}{}
	for _, v := range previous {
		addedByID[v.Id] = struct{}{}
		addedByText[t.getOriginalText(&v)] = struct{}{}
	}
	for _, v := range current {
		if _, ok := addedByID[v.Id]; ok {
			t.print(fmt.Sprintf("[twitter] found a duplicate (same id) from database id:%d, text:%s\n", v.Id, v.Text))
			continue
		}
		text := t.getOriginalText(&v)
		if _, ok := addedByText[text]; ok {
			t.print(fmt.Sprintf("[twitter] found a duplicate (same original text) from database id:%d, text:%s\n", v.Id, v.Text))
			continue
		}
		addedByID[v.Id] = struct{}{}
		addedByText[text] = struct{}{}
		diff = append(diff, v)
	}
	return diff
}

func (t *TwitterBot) removeDuplicates(current []anaconda.Tweet) []anaconda.Tweet {
	temp := map[string]struct{}{}
	stripped := []anaconda.Tweet{}
	for _, tweet := range current {
		text := t.getOriginalText(&tweet)
		if _, ok := temp[text]; !ok {
			temp[text] = struct{}{}
			stripped = append(stripped, tweet)
		} else {
			t.print(fmt.Sprintf("[twitter] found a duplicate (id:%d), text:%s\n", tweet.Id, tweet.Text))
		}
	}
	return stripped
}

func (t *TwitterBot) like(tweet *anaconda.Tweet) {
	if !t.likePolicy.auto {
		return
	}
	if tweet.FavoriteCount > t.likePolicy.threshold {
		_, err := t.twitterClient.Favorite(tweet.Id)
		if err != nil {
			t.print(fmt.Sprintf("[twitter] failed to like tweet (id:%d), error: %v\n", tweet.Id, err))
			return
		}
		log.Printf("[twitter] liked tweet (id:%d)\n", tweet.Id)
	} else if tweet.RetweetedStatus != nil &&
		tweet.RetweetedStatus.FavoriteCount > t.likePolicy.threshold {
		t.like(tweet.RetweetedStatus)
	}
}

func (t *TwitterBot) print(text string) {
	if t.debug {
		log.Println(text)
	}
}

func (t *TwitterBot) sleep() {
	if !t.debug {
		helpers.Sleep(maxRandTimeSleepBetweenRequests)
	}
}

func (t *TwitterBot) maybeSleep(chance, totalChance, min, max int) {
	if !t.debug {
		helpers.MaybeSleepMinMax(chance, totalChance, min, max)
	}
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

func (t *TwitterBot) unfollowUser(user *anaconda.User) {
	unfollowed, err := t.twitterClient.UnfollowUserId(user.Id)
	if err != nil {
		checkBotRestriction(err)
		t.print(fmt.Sprintf("[twitter] failed to unfollow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err))
	}
	log.Printf("[twitter] unfollowing user (id:%d, name:%s)\n", unfollowed.Id, unfollowed.Name)
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

func (t *TwitterBot) followUser(user *anaconda.User) {
	followed, err := t.twitterClient.FollowUserId(user.Id, nil)
	if err != nil && !checkUnableToFollowAtThisTime(err) {
		checkBotRestriction(err)
		t.print(fmt.Sprintf("[twitter] failed to follow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err))
	}
	log.Printf("[twitter] following user (id:%d, name:%s)\n", followed.Id, followed.Name)
}

// retweet retweets the first tweet been able to retweet.
// It returns an error if no retweet has been possible.
func (t *TwitterBot) retweet(current []anaconda.Tweet) (rt anaconda.Tweet, err error) {
	for _, tweet := range current {
		if t.retweetPolicy.like {
			t.like(&tweet)
		}
		retweet, err := t.twitterClient.Retweet(tweet.Id, false)
		if err != nil {
			t.print(fmt.Sprintf("[twitter] failed to retweet tweet (id:%d), error: %v\n", tweet.Id, err))
			t.followUser(&tweet.User)
			continue
		}
		rt = retweet
		if t.retweetPolicy.like {
			t.like(&rt)
		}
		log.Printf("[twitter] retweet (rid:%d, id:%d)\n", rt.Id, tweet.Id)
		t.followUser(&tweet.User)
		return rt, err
	}
	err = fmt.Errorf("unable to retweet")
	return rt, err
}

func (t *TwitterBot) getTweets(queries []string, previous []anaconda.Tweet) ([]anaconda.Tweet, error) {
	query := helpers.GetRandomElement(queries)
	log.Println("[twitter] searching tweets to retweet with query:", query)
	v := url.Values{}
	v.Set("count", strconv.Itoa(defaultMaxRetweetBySearch))
	results, err := t.twitterClient.GetSearch(query, v)
	if err != nil {
		return nil, err
	}
	current := results.Statuses
	current = t.removeDuplicates(current)
	current = t.takeDifference(previous, current)
	log.Println("[twitter] found", len(current), "tweet(s) to retweet matching pattern")
	return current, nil
}

func (t *TwitterBot) autoRetweet(queries []string) error {
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
			if count < t.retweetPolicy.maxTry {
				count++
				continue
			} else {
				return fmt.Errorf("[twitter] unable to retweet something after %d tries\n", t.retweetPolicy.maxTry)
			}
		}
		previous = append(previous, retweeted)
		tojson.Save(t.tweetsPath, previous)
		return nil
	}
}

func (t *TwitterBot) updateFollowers() error {
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

func (t *TwitterBot) updateFriends() error {
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

// unfollowFriend flags the friend as not followed anymore.
// We do not remove friends from database, we just flag them as non friend.
func (t *TwitterBot) unfollowFriend(id int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.friends.Ids[strconv.FormatInt(id, 10)].Follow = false
	err := tojson.Save(t.friendsPath, t.friends)
	if err != nil {
		log.Fatalln(err)
	}
}

func (t *TwitterBot) getFriendToUnFollow() (int64, bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for strID, user := range t.friends.Ids {
		// unfollow only if is followed and is in database from at least 1 day
		if time.Now().UnixNano()-user.Timestamp < oneDayInNano || !user.Follow {
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

func (t *TwitterBot) unfollowAll() {
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
	log.Println("[twitter] no more friends to unfollow, waiting 3 hours...")
	time.Sleep(3 * time.Hour)
	t.unfollowAll()
}

func (t *TwitterBot) isFollower(id int64) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	_, ok := t.followers.Ids[strconv.FormatInt(id, 10)]
	return ok
}

func (t *TwitterBot) getFriend(id int64) (*twitterUser, bool) {
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

func (t *TwitterBot) addFriend(id int64) {
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

func (t *TwitterBot) followAll(ids []int64) {
	for _, id := range ids {
		if _, ok := t.getFriend(id); ok || t.isFollower(id) {
			continue
		}
		user, err := t.twitterClient.FollowUserId(id, nil)
		if err != nil && !checkUnableToFollowAtThisTime(err) {
			checkBotRestriction(err)
			t.print(fmt.Sprintf("[twitter] failed to follow user (id:%d, name:%s), error: %v\n", user.Id, user.Name, err))
			continue
		}
		t.addFriend(id)
		log.Printf("[twitter] following (id:%d, name:%s)\n", user.Id, user.Name)
		time.Sleep(timeSleepBetweenFollowUnFollow)
		t.sleep()
		t.maybeSleep(1, 5, 5000, 10000)
	}
}

func (t *TwitterBot) fetchUserIds(query string, maxPage int) []int64 {
	users, err := t.twitterClient.GetUserSearch(query, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
	ids := []int64{}
	if len(users) == 0 {
		return nil
	}
	// gettings followers of the first user found
	user := users[0]
	nextCursor := "-1"
	currentPage := 1
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
