# GoMarkov
An attempt to parallelize a Markov Chain generator in Go.

A reworked version of [this official codewalk.](https://golang.org/doc/codewalk/markov/)

###Parameters

>-corpusFile string:
>       filename of the corpus file (default "corpus.txt")

>  -maxSentenceLength int:
>        maximum number of words in generated sentences (default 8)

>  -numSentences int:
>        number of sentences to generate (default 100)

>  -outputFile string:
>        filename of the ouput file (default "output.txt")

>  -prefixLen int:
>        prefix length in words (default 2)


Corpora are read from a single plaintext file.

