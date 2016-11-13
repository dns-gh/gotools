package main

import (
	"log"
	"math/rand"
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

func sleep(amount int) {
	random := rand.Intn(amount)
	log.Printf("random sleep: %+v seconds\n", random)
	time.Sleep(time.Second * time.Duration(random))
}

func getRandomElement(tab []string) string {
	return tab[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(tab))]
}

func trunc(text string) string {
	length := len(text)
	maxLength := 90
	if length > maxLength {
		length = maxLength
	}
	return text[0:length]
}
