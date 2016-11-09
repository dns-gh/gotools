package main

import (
	"os"

	"github.com/dns-gh/tojson"
)

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

func load() ([]object, error) {
	objects := &[]object{}
	if _, err := os.Stat("rocks.json"); os.IsNotExist(err) {
		save(*objects)
	}
	err := tojson.Load("rocks.json", objects)
	if err != nil {
		return nil, err
	}
	return *objects, nil
}

func save(objects []object) {
	tojson.Save("rocks.json", &objects)
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

func update(current []object) ([]object, error) {
	previous, err := load()
	if err != nil {
		return nil, err
	}
	merged, diff := merge(previous, current)
	save(merged)
	return diff, nil
}
