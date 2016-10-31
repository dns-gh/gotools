package algorithms

import (
	"fmt"
	"runtime"
	"sync"
)

type state struct {
	charset   string
	indexes   []int
	candidate string
	maxLength int
	mutex     sync.Mutex
	done      chan struct{}
	quit      sync.WaitGroup
}

func replaceAt(str string, replacement string, index int) string {
	return str[:index] + replacement + str[index+1:]
}

func makestate(charset string, maxLength int) *state {
	return &state{
		charset:   charset,
		candidate: "",
		indexes:   make([]int, maxLength),
		maxLength: maxLength,
		done:      make(chan struct{}),
	}
}

func (s *state) stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.done == nil {
		return
	}
	close(s.done)
	s.done = nil
}

func (s *state) incremente(index int) {
	if index >= s.maxLength {
		s.candidate += "_invalid"
		return
	}
	if s.indexes[index]+1 == len(s.charset) {
		s.indexes[index] = 0
		s.candidate = replaceAt(s.candidate, s.charset[0:1], index)
		s.incremente(index + 1)
		return
	}
	if index >= len(s.candidate) {
		s.candidate += " "
	} else {
		s.indexes[index] += 1
	}
	at := s.indexes[index]
	s.candidate = replaceAt(s.candidate, s.charset[at:at+1], index)
}

func (s *state) next() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.candidate == "" {
		s.candidate = s.charset[0:1]
		return s.candidate
	}
	s.incremente(0)
	return s.candidate
}

// BruteForce launch brute-force algorithm using given input and checks
// if there is a match using the operand argument
func BruteForce(maxLength, cpu int, charset string, operand func(string) bool) error {
	runtime.GOMAXPROCS(cpu)
	state := makestate(charset, maxLength)

	solution := ""
	for i := 0; i < cpu; i++ {
		state.quit.Add(1)
		go func() {
			for {
				select {
				case <-state.done:
					state.quit.Done()
					return
				default:
					candidate := state.next()
					if len(candidate) > state.maxLength {
						state.stop()
						state.quit.Done()
						return
					}
					if operand(candidate) {
						solution = candidate
						state.stop()
						state.quit.Done()
						return
					}
				}
			}
		}()
	}
	state.quit.Wait()
	if len(solution) == 0 {
		return fmt.Errorf("brute force failed to find a proper candidate")
	}
	return nil
}
