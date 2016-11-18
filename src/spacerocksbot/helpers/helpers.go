package helpers

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

func QuickSort(values []int64) {
	sort(values, 0, len(values)-1)
}

func ParseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func ParseInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func ParseTime(value string, timeFormat string) time.Time {
	parsed, err := time.Parse(timeFormat, value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func ParseDuration(value string) time.Duration {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalln(err)
	}
	return parsed
}

func getRandom(duration int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(duration)
}

func Sleep(duration int) {
	random := getRandom(duration)
	time.Sleep(time.Second * time.Duration(random))
}

func SleepMinMax(min, max int) {
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

func MaybeSleepMinMax(chance, totalChance, min, max int) {
	if maybe(chance, totalChance) {
		SleepMinMax(min, max)
	}
}

func maybe(chance, totalChance int) bool {
	random := getRandom(totalChance)
	if random <= chance {
		return true
	}
	return false
}

func GetRandomElement(tab []string) string {
	return tab[getRandom(len(tab))]
}

// truncate text to a hardcoded length.
// It helps to visualize logs in the console.
func Trunc(text string) string {
	length := len(text)
	maxLength := 90
	if length > maxLength {
		length = maxLength
	}
	return text[0:length]
}
