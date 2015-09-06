package main

import (
	"flag"
	"github.com/Jkolios/GoMarkov/markov"
	"github.com/Jkolios/GoMarkov/myRedis"
	"log"
	"math/rand"
	"os"
	"time"
)

func buildChain(corpusFilename string, prefixLen int) (*markov.Chain, error) {
	corpusFP, err := os.Open(corpusFilename)
	if err != nil {
		log.Println("Cannot open corpus file")
		return nil, err
	}
	defer func() {
		if err := corpusFP.Close(); err != nil {
			panic(err)
		}
	}()

	chain := markov.NewChain(prefixLen)
	chain.Build(corpusFP)
	return chain, nil
}

func producer(chain *markov.Chain, redisConnection *myRedis.RedisState, maxSentenceLength int, sync, control chan bool) {

	log.Println("In producer")

	defer func(control chan bool) { control <- true }(control)
	state := false

	for {
		select {
		case newState := <-sync:
			state = newState
			log.Printf("producer received %v on sync channel\n", newState)
		case <-control:
			return
		default:
			if state == true {
				if !redisConnection.Low() {
					log.Printf("producer activated the consumer\n")
					sync <- true
				}
				if redisConnection.Full() {
					log.Printf("producer determined the queue is full\n")
					state = false
					sync <- true
					log.Println("Producer self halted: queue full")
					continue
				}

				chain.Generate(10, maxSentenceLength, redisConnection)

			}
		}

	}
}

func consumer(redisConnection *myRedis.RedisState, sync, control chan bool, output chan string) {

	log.Println("In consumer")

	defer func(control chan bool) { control <- true }(control)
	state := false

	for {
		select {
		case newState := <-sync:
			state = newState
			log.Printf("consumer received %v on sync channel\n", newState)

		case <-control:
			return

		default:
			if state == true {
				if !redisConnection.Full() {
					log.Printf("consumer activated the producer\n")
					sync <- true
				}
				if redisConnection.Low() {
					log.Printf("consumer determined the queue is low\n")
					state = false
					sync <- true
					log.Println("Consumer self halted: queue low")
					continue
				}

				outputStr, err := redisConnection.GetString()
				if err != nil {
					output <- err.Error()
				} else {
					output <- outputStr
				}

			}

		}
	}
}

func main() {
	corpusfile := flag.String("corpusFile", "corpus.txt", "filename of the corpus file")
	maxSentenceLength := flag.Int("maxSentenceLength", 8, "maximum number of words in generated sentences")
	prefixLen := flag.Int("prefixLen", 2, "prefix length in words")
	redisHost := flag.String("redisHost", "localhost", "host of the redis instance")
	redisPort := flag.String("redisPort", "6379", "port of the redis instance")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	chain, err := buildChain(*corpusfile, *prefixLen)
	if err != nil {
		log.Println("Cannot build Markov chain from given input")
		return
	}

	// Redis connection pool
	redis := myRedis.InitRedis(*redisHost, *redisPort)

	syncChannel := make(chan bool, 0)
	controlChannel := make(chan bool, 0)
	outputChannel := make(chan string, 0)

	go producer(chain, redis, *maxSentenceLength, syncChannel, controlChannel)
	go consumer(redis, syncChannel, controlChannel, outputChannel)
	syncChannel <- true

	runningGoroutines := 2

Runloop:
	for {
		select {
		case <-controlChannel:
			log.Println("Received goroutine termination signal")
			runningGoroutines--
			if runningGoroutines == 0 {
				break Runloop
			}
		case output := <-outputChannel:
			log.Printf("Output: %v\n", output)
		}
	}

	log.Println("All goroutines have terminated, quitting.")

}
