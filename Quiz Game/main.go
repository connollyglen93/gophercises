package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type round struct {
	q string //question
	a string //answer
	s string //solution
}

type game struct {
	rounds []round
}

var wg sync.WaitGroup

func main() {
	var timeLimit int
	var shuffle bool
	flag.IntVar(&timeLimit, "time", 10, "Define how long the user had to complete the quiz")
	flag.BoolVar(&shuffle, "shuffle", false, "Define whether the quiz is shuffled (default is false)")
	flag.Parse()

	game, err := getRounds("problems.csv", shuffle)
	if err != nil {
		panic(err)
	}

	console := bufio.NewReader(os.Stdin)
	strLimit := strconv.Itoa(timeLimit)

	fmt.Printf("You have %s seconds to answer %d questions...\n", strLimit, len(game.rounds))
	fmt.Println("Press Any Key to Begin...")
	console.ReadByte()

	wg.Add(len(game.rounds))
	//Purposely utilizing the fact that the `game` variable will change it's value after the declaration of this function.
	go func() {
		duration, err := time.ParseDuration(strLimit + "s")
		if err != nil {
			panic(err)
		}
		time.Sleep(duration)
		fmt.Println("Times Up!")
		game.printSummary()
		os.Exit(1)
	}()

	for i := range game.rounds {
		processRound(&game, i)
	}
	fmt.Println("Quiz Completed!")
	game.printSummary()
}

func (game game) printSummary() {

	var (
		correct   int
		incorrect int
		score     float64
	)

	for _, round := range game.rounds {
		if round.isCorrect() {
			correct++
		} else {
			incorrect++
		}
	}
	if correct > 0 {
		score = (float64(correct) / float64(len(game.rounds))) * 100.0
	} else {
		score = 0
	}

	fmt.Printf("Score:\nCorrect: %d\nIncorrect: %d\nTotalScore: %.1f%% \n", correct, incorrect, score)
}

func (game *game) shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(game.rounds), func(i, j int) { game.rounds[i], game.rounds[j] = game.rounds[j], game.rounds[i] })
}

func processRound(game *game, i int) {
	console := bufio.NewReader(os.Stdin)

	fmt.Println(game.rounds[i].q + ":")

	solution, err := console.ReadString('\n')
	if err != nil {
		panic(err)
	}
	solution = strings.ToLower(strings.TrimSpace(solution))
	game.rounds[i].setSolution(solution)
	result := "Incorrect!"
	if game.rounds[i].isCorrect() {
		result = "Correct!"
	}
	fmt.Println(result + "\n")
}

func (r round) isCorrect() bool {
	return strings.Compare(r.s, r.a) == 0
}

func (r *round) setSolution(solution string) {
	r.s = solution
}

func getRounds(path string, shuffle bool) (game, error) {
	f, err := os.Open(path)
	game := game{make([]round, 0)}
	if err != nil {
		return game, err
	}
	defer f.Close()

	ls, err := csv.NewReader(f).ReadAll()

	if err != nil {
		return game, err
	}

	for _, line := range ls {
		var round round
		round.q = strings.TrimSpace(line[0])
		round.a = strings.ToLower(strings.TrimSpace(line[1]))
		game.rounds = append(game.rounds, round)
	}

	if shuffle {
		game.shuffle()
	}

	return game, nil
}
