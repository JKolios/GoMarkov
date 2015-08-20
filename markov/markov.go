package markov

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"runtime"
	"sync"
)

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
	writeMutex sync.Mutex
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen, sync.Mutex{}}
}

func readerToWords(r io.ReadSeeker) ([]string, int) {
	var outWords []string
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	count := 0
	for scanner.Scan() {
		outWords = append(outWords, scanner.Text())
		count++
	}
	return outWords, count
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.ReadSeeker) {
	words, wordCount := readerToWords(r)

	goroutineCount := runtime.GOMAXPROCS(0)

	syncChan := make(chan bool)
	for i := 0; i<goroutineCount; i++ {
		go c.chainBuilder(words[i * (wordCount/goroutineCount): (i+1) * (wordCount/goroutineCount)], syncChan)
		fmt.Printf("Launching goroutine number %v with limits: %v %v\n", i, i * wordCount/goroutineCount, (i+1) * wordCount/goroutineCount)
	}

	for i := 0; i<goroutineCount; i++ {
		<-syncChan
	}
	fmt.Println("All goroutines returned\n")
}

func(c *Chain) chainBuilder(words []string, syncChan chan bool) {
	p := make(Prefix, c.prefixLen)
		for _, word := range words {
			key := p.String()
			c.writeMutex.Lock()
			c.chain[key] = append(c.chain[key], word)
			c.writeMutex.Unlock()
			p.Shift(word)
	}
	syncChan<-true
}

// Generate writes nStrings strings of at most nWords each to writer. The strings are generated from Chain.
func (c *Chain) Generate(nStrings, nWords int, writer io.Writer) {
	p := make(Prefix, c.prefixLen)
	for outString := 0; outString < nStrings; outString++ {
		for i := 0; i < nWords; i++ {
			choices := c.chain[p.String()]
			if len(choices) == 0 {
				break
			}
			next := choices[rand.Intn(len(choices))]
			writer.Write([]byte(next + " "))
			p.Shift(next)
		}
		writer.Write([]byte("\n"))
	}
}

// GetChain returs the generated chain
func (c *Chain) GetChain() map[string][]string {
	return c.chain
}
