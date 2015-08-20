package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Jkolios/GoMarkov/markov"
	"math/rand"
	"os"
	"time"
)

func main() {
	corpusfile := flag.String("corpusFile", "corpus.txt", "filename of the corpus file")
	outputFile := flag.String("outputFile", "output.txt", "filename of the ouput file")
	numSentences := flag.Int("numSentences", 100, "number of sentences to generate")
	maxSentenceLength := flag.Int("maxSentenceLength", 8, "maximum number of words in generated sentences")
	prefixLen := flag.Int("prefixLen", 2, "prefix length in words")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	corpusFP, err := os.Open(*corpusfile)
	if err != nil {
		fmt.Println("Cannot open corpus file")
		return
	}
	defer func() {
		if err := corpusFP.Close(); err != nil {
			panic(err)
		}
	}()

	chain := markov.NewChain(*prefixLen)
	chain.Build(corpusFP)

	outputFP, err := os.Create(*outputFile)
	if err != nil {
		fmt.Println("Cannot open output file")
		return
	}
	defer func() {
		if err := outputFP.Close(); err != nil {
			panic(err)
		}
	}()

	writer := bufio.NewWriter(outputFP)
	chain.Generate(*numSentences, *maxSentenceLength, writer)
	writer.Flush()

	fmt.Println("Output successfully written.")

}
