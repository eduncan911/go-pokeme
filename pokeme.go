package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

// Flags
var (
	// F_GOROUTINES represents the number of GoRoutines to use
	F_GOROUTINES int
	// F_ROUNDS represents the total number of hands played by the group of players
	F_ROUNDS int
	// F_PLAYERS represents the number of players within the group
	F_PLAYERS int
)

var (
	bufferOfWork    int
	delayRandomizer = rand.New(rand.NewSource(time.Now().UnixNano()))
	cardRandomizer  = rand.New(rand.NewSource(time.Now().UnixNano()))

	in  chan int
	out chan *Results
)

// FlopTurnRiver represents the community cards
type FlopTurnRiver struct {
	Card1, Card2, Card3, Card4, Card5 int
}

// PlayerResult represents a single player's results in the round played
type PlayerResult struct {
	Card1, Card2 int
	Folded       bool

	// add more properties here
	//
}

// Results will hold the results from each round played
type Results struct {
	FlopTurnRiver *FlopTurnRiver
	PlayerResults []*PlayerResult
}

// WorkOnRound takes in a channel of ints (e.g. max rounds to be played),
// iterates over them, and outputs their results on the out channel
func WorkOnRound(in <-chan int, out chan<- *Results) {

	fmt.Println("WorkOnRound created")

	// since we are blocking on "receiving" data in the "in" parameter,
	// this externalizes the buffering to whatever is sending to us.
	//
	// it executes each for loop asynchronously, as things are pushed onto
	// the channel.
	//
	// it also "pauses" and waits for something to be sent on the channel.
	//
	for i := range in {

		fmt.Println("WorkOnRound starting:", i)

		flopTurnRiver := &FlopTurnRiver{
			Card1: cardRandomizer.Intn(13),
			Card2: cardRandomizer.Intn(13),
			Card3: cardRandomizer.Intn(13),
			Card4: cardRandomizer.Intn(13),
			Card5: cardRandomizer.Intn(13),
		}
		playerResults := make([]*PlayerResult, F_PLAYERS)

		//
		// do work here...
		//
		//d := time.Millisecond * time.Duration(delayRandomizer.Intn(2000))
		//time.Sleep(d)
		//
		//
		//

		for i := 0; i < F_PLAYERS; i++ {
			playerResults[i] = &PlayerResult{
				Card1: cardRandomizer.Intn(13),
				Card2: cardRandomizer.Intn(13),
				//Folded: (cardRandomizer.Intn(2)%2 == 0),
			}
		}

		// results of work done
		r := &Results{
			FlopTurnRiver: flopTurnRiver,
			PlayerResults: playerResults,
		}

		fmt.Println("WorkOnRound stopped:", i)

		// push results onto channel
		out <- r
	}
}

// ProcessHands is what sends the massive amount of hands (e.g. 1,000,000) to
// to be processed.  Since FlopTurnRiver channel is buffered, it will block on
// any new submissions until the channel frees up.
func ProcessHands(c chan int) {
	for i := 0; i < F_ROUNDS; i++ {
		c <- i
	}
}

// ProcessResults receives results from a channel and processes them.
func ProcessResults(c chan *Results) {

	// this will sit and wait for a response
	for range c {
		// fmt.Printf("Results: %#v\n", r.FlopTurnRiver)
		// for p := range r.PlayerResults {
		// 	fmt.Printf("Player: %#v\n", p)
		// }
	}
}

func main() {

	// since this app is very CPU dependent (no external processes blocking),
	// limit things to 4x
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	// externalize flag validation
	if err := parseFlags(); err != nil {
		fmt.Println(err.Error())
		fmt.Println("Use -h for help")
		return
	}

	fmt.Println("Starting routines")
	in = make(chan int, bufferOfWork)
	out = make(chan *Results, bufferOfWork)

	// create a fixed number concurrent goroutines to 'buffer'.
	// each gorountine sits and "waits" for input.
	//
	// note this does not block, and can continue to ProcessHands
	// and ProcessResults below
	//
	for i := 1; i <= F_GOROUTINES; i++ {
		fmt.Println("building routine ", i)
		go WorkOnRound(in, out)
	}

	// this is what sends the inputs to the above routines
	go ProcessHands(in)

	// ProcessResults blocks, waiting for results to process
	ProcessResults(out)
}

func parseFlags() error {
	flag.IntVar(&F_GOROUTINES, "g", 1000, "Number of concurrent goroutines")
	flag.IntVar(&F_ROUNDS, "r", 10000, "Number of rounds played for the group")
	flag.IntVar(&F_PLAYERS, "p", 6, "Number of players per deal")
	flag.Parse()

	// flag validation examples
	if F_GOROUTINES < 1 {
		return fmt.Errorf("Parameter 'g' must be 1 or larger.  It was: %d", F_PLAYERS)
	}

	if F_ROUNDS < 1 {
		return fmt.Errorf("Parameter 'r' must be 1 or larger.  It was: %d", F_PLAYERS)
	}

	if F_PLAYERS < 2 {
		return fmt.Errorf("Parameter 'p' must be 2 or larger.  It was: %d", F_PLAYERS)
	}

	if F_ROUNDS < F_GOROUTINES {
		return fmt.Errorf("Parameter 'r' must be less than and divisible by 'g' as a whole number. For example: -r=10000 -g=10 would yield 10000 / 10, which means 10 GoRoutines will process 1000 hands at a time.")
	}

	// TODO: capture divisible-by-0 error here (panic)
	bufferOfWork = F_ROUNDS / F_GOROUTINES

	return nil
}
