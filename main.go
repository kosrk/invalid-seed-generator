package main

import (
	"brute/pkg/seed"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"
)

var targetPubkey, _ = hex.DecodeString("your_pubkey_here_in_hex_without_0x")

const (
	get12Words  = false
	brute2Words = false
)

func main() {
	seedWords := strings.Split("your_seed_here", " ")
	words := seed.WordsArr

	if len(seedWords) != 12 && len(seedWords) != 24 {
		panic("seed should have 12 or 24 words")
	}
	for _, s := range seedWords {
		if !seed.Words[s] {
			panic("unknown word in seed: " + s)
		}
	}

	seedCh := make(chan string, 100_000)
	for i := 0; i < 50_000; i++ {
		go worker(targetPubkey, seedCh)
	}

	if get12Words {
		seedWords = seedWords[:12]
	}

	var iters int
	if brute2Words {
		iters = bruteforce2Words(seedWords, words, seedCh)
	} else {
		iters = bruteforce1Words(seedWords, words, seedCh)
	}
	close(seedCh)
	log.Printf("waiting 10s...\n")
	time.Sleep(10 * time.Second)
	log.Printf("Processed: %d but seed not found\n", iters)
}

func bruteforce1Words(seedWords []string, words []string, seedChan chan string) int {
	totalPossible := len(seedWords) * len(words)
	log.Printf("One word. Total possible variations: %d\n", totalPossible)
	iteration := 0
	for i := 0; i < len(seedWords); i++ {
		for _, word1 := range words {
			newWords := make([]string, len(seedWords))
			copy(newWords, seedWords)
			newWords[i] = word1
			seedChan <- strings.Join(newWords, " ")
		}
		iteration += len(words)
		log.Printf("Iters: %d Progress: %.2f%s\n", iteration, 100*float64(iteration)/float64(totalPossible), "%")
	}
	return iteration
}

func bruteforce2Words(seedWords []string, words []string, seedChan chan string) int {
	totalPossible := len(seedWords) * (len(seedWords) - 1) / 2 * len(words) * len(words)
	log.Printf("Two words. Total possible variations: %d\n", totalPossible)
	iteration := 0
	step := 100 * len(words)
	start := time.Now()
	for i := 0; i < len(seedWords); i++ {
		log.Printf("word1 position: %d\n", i)
		for j := i + 1; j < len(seedWords); j++ {
			for _, word1 := range words {
				for _, word2 := range words {
					newWords := make([]string, len(seedWords))
					copy(newWords, seedWords)
					newWords[i] = word1
					newWords[j] = word2
					seedChan <- strings.Join(newWords, " ")
				}
				iteration += len(words)
				if iteration%(step) == 0 {
					t := time.Since(start)
					speed := t / time.Duration(iteration)
					leftTime := speed * time.Duration(totalPossible-iteration)
					log.Printf("Iters: %d Progress: %.2f%s estimated time: %s\n", iteration, 100*float64(iteration)/float64(totalPossible), "%", leftTime)
				}
			}
		}
	}
	return iteration
}

func worker(targetPubkey []byte, seedChan chan string) {
	for x := range seedChan {
		pk, err := seed.ToPrivateKeyBip39(x)
		if err != nil {
			panic(err)
		}
		pubkey := pk.Public().(ed25519.PublicKey)
		if slices.Equal(pubkey, targetPubkey) {
			fmt.Printf("Found seed: %v\n", x)
			os.Exit(0)
		}
	}
}
