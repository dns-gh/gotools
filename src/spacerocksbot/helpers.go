package main

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
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

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func parseInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func parseTime(value string, timeFormat string) time.Time {
	parsed, err := time.Parse(timeFormat, value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func parseDuration(value string) time.Duration {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func getRandom(duration int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(duration)
}

func sleep(duration int) {
	random := getRandom(duration)
	time.Sleep(time.Second * time.Duration(random))
}

func sleepMinMax(min, max int) {
	if max < min {
		temp := min
		min = max
		max = temp
	} else if max == min {
		max = min + 1
	}
	random := getRandom(max - min)
	time.Sleep(time.Second * time.Duration(min+random))
}

func maybeSleepMinMax(chance, totalChance, min, max int) {
	if maybe(chance, totalChance) {
		sleepMinMax(min, max)
	}
}

func maybe(chance, totalChance int) bool {
	random := getRandom(totalChance)
	if random <= chance {
		return true
	}
	return false
}

func getRandomElement(tab []string) string {
	return tab[getRandom(len(tab))]
}

// truncate text to a hardcoded length.
// It helps to visualize logs in the console.
func trunc(text string) string {
	length := len(text)
	maxLength := 90
	if length > maxLength {
		length = maxLength
	}
	return text[0:length]
}

func makeLog(path string) (string, *os.File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	err = os.MkdirAll(filepath.Dir(abs), os.ModePerm)
	if err != nil {
		return "", nil, err
	}
	f, err := os.OpenFile(abs, os.O_WRONLY+os.O_APPEND+os.O_CREATE, os.ModePerm)
	return abs, f, err
}
