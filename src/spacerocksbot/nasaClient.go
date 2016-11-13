package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	conf "github.com/dns-gh/flagsconfig"
	"github.com/dns-gh/tojson"
)

type nasaClient struct {
	apiKey      string
	firstOffset int
	offset      int
	poll        time.Duration
	path        string
	body        string // orbiting body to watch
	debug       bool
}

func (n *nasaClient) hasDefaultKey() bool {
	return n.apiKey == nasaAPIDefaultKey
}

func makeNasaClient(config *conf.Config) *nasaClient {
	apiKey := os.Getenv("NASA_API_KEY")
	if len(apiKey) == 0 {
		apiKey = nasaAPIDefaultKey
	}
	return &nasaClient{
		apiKey:      apiKey,
		firstOffset: parseInt(config.Get(firstOffsetFlag)),
		offset:      parseInt(config.Get(offsetFlag)),
		poll:        parseDuration(config.Get(pollFrequencyFlag)),
		path:        config.Get(nasaPathFlag),
		debug:       parseBool(config.Get(debugFlag)),
	}
}

type links struct {
	Next string `json:"next"`
	Prev string `json:"prev"`
	Self string `json:"self"`
}

type diameter struct {
	EstimatedDiameterMin float64 `json:"estimated_diameter_min"`
	EstimatedDiameterMax float64 `json:"estimated_diameter_max"`
}

type estimatedDiameter struct {
	Kilometers diameter `json:"kilometers"`
	Meters     diameter `json:"meters"`
	Miles      diameter `json:"miles"`
	Feet       diameter `json:"feet"`
}

type relativeVelocity struct {
	KilometersPerSecond string `json:"kilometers_per_second"`
	KilometersPerHour   string `json:"kilometers_per_hour"`
	MilesPerHour        string `json:"miles_per_hour"`
}

type missDistance struct {
	Astronomical string `json:"astronomical"`
	Lunar        string `json:"lunar"`
	Kilometers   string `json:"kilometers"`
	Miles        string `json:"miles"`
}

type closeApprochInfo struct {
	CloseApproachDate      string           `json:"close_approach_date"`
	EpochDateCloseApproach int64            `json:"epoch_date_close_approach"`
	RelativeVelocity       relativeVelocity `json:"relative_velocity"`
	MissDistance           missDistance     `json:"miss_distance"`
	OrbitingBody           string           `json:"orbiting_body"`
}

type object struct {
	Links                          links              `json:"links"`
	NeoReferenceID                 string             `json:"neo_reference_id"`
	Name                           string             `json:"name"`
	NasaJplURL                     string             `json:"nasa_jpl_url"`
	AbsoluteMagnitudeH             float64            `json:"absolute_magnitude_h"`
	EstimatedDiameter              estimatedDiameter  `json:"estimated_diameter"`
	IsPotentiallyHazardousAsteroid bool               `json:"is_potentially_hazardous_asteroid"`
	CloseApproachData              []closeApprochInfo `json:"close_approach_data"`
}

// SpaceRocks (asteroids) represents all asteroids data available between two dates.
// The information is stored in the NearEarthObjects map.
// [Generated with the help of https://mholt.github.io/json-to-go/]
type SpaceRocks struct {
	Links        links `json:"links"`
	ElementCount int   `json:"element_count"`
	// the key of the NearEarthObjects map represents a date with the following format YYYY-MM-DD
	NearEarthObjects map[string][]object `json:"near_earth_objects"`
}

func (n *nasaClient) load() ([]object, error) {
	objects := &[]object{}
	if _, err := os.Stat(n.path); os.IsNotExist(err) {
		tojson.Save(n.path, objects)
	}
	err := tojson.Load(n.path, objects)
	if err != nil {
		return nil, err
	}
	return *objects, nil
}

func merge(previous, current []object) ([]object, []object) {
	merged := []object{}
	diff := []object{}
	added := map[string]struct{}{}
	for _, v := range previous {
		added[v.NeoReferenceID] = struct{}{}
		merged = append(merged, v)
	}
	for _, v := range current {
		if _, ok := added[v.NeoReferenceID]; ok {
			continue
		}
		added[v.NeoReferenceID] = struct{}{}
		merged = append(merged, v)
		diff = append(diff, v)
	}
	return merged, diff
}

func (n *nasaClient) update(current []object) ([]object, error) {
	previous, err := n.load()
	if err != nil {
		return nil, err
	}
	merged, diff := merge(previous, current)
	tojson.Save(n.path, merged)
	return diff, nil
}

func (n *nasaClient) fetchRocks(days int) (*SpaceRocks, error) {
	if days > 7 {
		return nil, fmt.Errorf(fetchMaxSizeError)
	} else if days < -7 {
		return nil, fmt.Errorf(fetchMaxSizeError)
	}
	now := time.Now()
	start := ""
	end := ""
	if days >= 0 {
		start = now.Format(nasaTimeFormat)
		end = now.AddDate(0, 0, days).Format(nasaTimeFormat)
	} else {
		start = now.AddDate(0, 0, days).Format(nasaTimeFormat)
		end = now.Format(nasaTimeFormat)
	}
	url := nasaAsteroidsAPIGet +
		n.apiKey +
		"&start_date=" + start +
		"&end_date=" + end
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(bytes), "OVER_RATE_LIMIT") {
		return nil, fmt.Errorf("http get rate limit reached, wait of use a proper key instead of the default one")
	}

	spacerocks := &SpaceRocks{}
	json.Unmarshal(bytes, spacerocks)
	return spacerocks, nil
}

func (n *nasaClient) getDangerousRocks(offset int) ([]object, error) {
	rocks, err := n.fetchRocks(offset)
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
						object.CloseApproachData[0].OrbitingBody == n.body {
						t := parseTime(object.CloseApproachData[0].CloseApproachDate, nasaTimeFormat)
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

func (n *nasaClient) sleep() {
	if !n.debug {
		sleep(maxRandTimeSleepBetweenTweets)
	}
}

func (n *nasaClient) fetch(offset int) ([]string, error) {
	log.Println("[nasa] checking nasa rocks...")
	current, err := n.getDangerousRocks(offset)
	if err != nil {
		return nil, err
	}
	log.Println("[nasa] found", len(current), "potential rocks to tweet")
	// TODO only merge and save asteroids once they are tweeted ?
	diff, err := n.update(current)
	if err != nil {
		return nil, err
	}
	formatedDiff := []string{}
	for _, object := range diff {
		sleep(maxRandTimeSleepBetweenTweets)
		closeData := object.CloseApproachData[0]
		approachDate := parseTime(closeData.CloseApproachDate, nasaTimeFormat)
		// extract lisible name
		name := object.Name
		parts := strings.SplitN(object.Name, " ", 2)
		if len(parts) == 2 {
			name = parts[1]
		}
		// extract lisible speed
		speed := closeData.RelativeVelocity.KilometersPerSecond
		parts = strings.Split(speed, ".")
		if len(parts) == 2 && len(parts[1]) > 2 {
			speed = parts[0] + "." + parts[1][0:1]
		}
		// extract lisible month
		month := approachDate.Month().String()
		if len(month) >= 3 {
			month = month[0:3]
		}
		// build status message
		statusMsg := fmt.Sprintf("A #%s #asteroid %s, Ø ~%.2f km and ~%s km/s is coming close to #%s on %s. %02d (details here %s)",
			getRandomElement(asteroidsQualificativeAdjective),
			name,
			(object.EstimatedDiameter.Kilometers.EstimatedDiameterMin+object.EstimatedDiameter.Kilometers.EstimatedDiameterMax)/2,
			speed,
			n.body,
			month,
			approachDate.Day(),
			object.NasaJplURL)
		formatedDiff = append(formatedDiff, statusMsg)
	}
	return formatedDiff, nil
}
