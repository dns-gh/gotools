// Space Rocks Watch is a bot watching
// asteroids coming too close to earth for the incoming week.
// It uses the nasa public API: https://api.nasa.gov/api.html
// You must set the NASA_API_KEY environement variable when
// launching the bot.
// You can get one here: https://api.nasa.gov/index.html#apply-for-an-api-key
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	nasaAsteroidsAPIGet = "https://api.nasa.gov/neo/rest/v1/feed?api_key="
	timeFormat          = "2006-01-02"
	orbitingBodyToWatch = "Earth"
)

func sort(tab []int64, left int, right int) {
	if left >= right {
		return
	}
	pivot := tab[left]
	i := left + 1
	for j := left; j <= right; j++ {
		if pivot > tab[j] {
			tab[i], tab[j] = tab[j], tab[i]
			i++
		}
	}
	tab[left], tab[i-1] = tab[i-1], pivot
	sort(tab, left, i-2)
	sort(tab, i, right)
}

func quickSort(values []int64) {
	sort(values, 0, len(values)-1)
}

func fetchRocks(days int) (*SpaceRocks, error) {
	if days > 7 {
		return nil, fmt.Errorf("cannot fetch infos for more than 7 days in one request")
	}
	now := time.Now()
	url := nasaAsteroidsAPIGet +
		os.Getenv("NASA_API_KEY") +
		"&start_date=" + now.Format(timeFormat) +
		"&end_date=" + now.AddDate(0, 0, days).Format(timeFormat)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer resp.Body.Close()

	spacerocks := &SpaceRocks{}
	json.NewDecoder(resp.Body).Decode(spacerocks)
	return spacerocks, nil
}

func parseTime(date string) (time.Time, error) {
	return time.Parse(timeFormat, date)
}

func filterRocks(spacerocks *SpaceRocks) ([]object, error) {
	dangerous := map[int64]object{}
	keys := []int64{}
	for _, v := range spacerocks.NearEarthObjects {
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

func main() {
	spacerocks, err := fetchRocks(7)
	if err != nil {
		log.Fatalln(err.Error())
	}
	objects, err := filterRocks(spacerocks)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println("+--------------------------------------------------------------+")
	fmt.Println("|               Potential dangerous incoming rocks             |")
	fmt.Println("+--------------------------------------------------------------+")
	for _, object := range objects {
		t, err := parseTime(object.CloseApproachData[0].CloseApproachDate)
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("| Diameter [%.2f to %.2f] km / coming near %s on %d-%02d-%02d |\n",
			object.EstimatedDiameter.Kilometers.EstimatedDiameterMin,
			object.EstimatedDiameter.Kilometers.EstimatedDiameterMax,
			orbitingBodyToWatch,
			t.Year(), t.Month(), t.Day())
	}
	fmt.Println("+--------------------------------------------------------------+")
}
