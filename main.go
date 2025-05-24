package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	ColorReset  = "\u001B[0m"
	ColorGreen  = "\u001B[32m"
	ColorYellow = "\u001B[33m"
	ColorWhite  = "\u001B[37m"
)

const (
	wordListFile       = "wordle-words.txt"
	statsFile          = "stats.csv"
	maxAttempts        = 6
	wordLength         = 5
	allPossibleLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type LetterFeedback struct {
	Char   rune
	Status string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please provide a number as command line argument")
		return
	}

	wordIndexArg := os.Args[1]
	wordIndex, err := strconv.Atoi(wordIndexArg)
	if err != nil {
		fmt.Println("Invalid command-line argument. Please launch with a valid number.")
		return
	}

	allWords, errLoadWords := loadWords(wordListFile)
	var secretWord string
	validGameSetup := false

	if errLoadWords == nil && allWords != nil {
		if wordIndex >= 0 && wordIndex < len(allWords) {
			secretWord = allWords[wordIndex]
			validGameSetup = true
		}
	}

	gameInputScanner := bufio.NewScanner(os.Stdin)
	var username string

	fmt.Print("Enter your username: ")

	if !validGameSetup {
		fmt.Print("Invalid word number.\n")
		fmt.Print("Press Enter to exit...\n")
		if gameInputScanner.Scan() {
		}
		return
	}

	eofDuringUsername := false
	if gameInputScanner.Scan() {
		username = strings.TrimSpace(gameInputScanner.Text())
	} else {
		eofDuringUsername = true
	}

	fmt.Print("Welcome to Wordle! Guess the 5-letter word.\n")

	attemptsLeft := maxAttempts
	var gameWon bool
	guessedIncorrectLetters := make(map[rune]bool)
	eofDuringGuessLoop := false

	if !eofDuringUsername {
		for attemptsLeft > 0 {
			fmt.Print("Enter your guess:  ")
			if !gameInputScanner.Scan() {
				eofDuringGuessLoop = true
				break
			}
			guess := strings.ToLower(strings.TrimSpace(gameInputScanner.Text()))

			if len(guess) != wordLength {
				fmt.Print("Your guess must be exactly 5 letters long.\n")
				continue
			}

			onlyLowercase := true
			for _, char := range guess {
				if char < 'a' || char > 'z' {
					onlyLowercase = false
					break
				}
			}
			if !onlyLowercase {
				fmt.Print("Your guess must only contain lowercase letters.\n")
				continue
			}

			if !isWordInList(guess, allWords) {
				fmt.Print("Word not in list. Please enter a valid word.\n")
				continue
			}

			attemptsLeft--
			detailedFeedback, currentGuessWin := generateFeedback(guess, secretWord)

			if currentGuessWin {
				gameWon = true
				fmt.Print("Congratulations! You've guessed the word correctly.\n")
				break
			}

			secretWordRunes := []rune(secretWord)
			for _, fb := range detailedFeedback {
				isCharInSecret := false
				for _, secretRune := range secretWordRunes {
					if unicode.ToLower(fb.Char) == unicode.ToLower(secretRune) {
						isCharInSecret = true
						break
					}
				}
				if fb.Status == "white" && !isCharInSecret {
					guessedIncorrectLetters[unicode.ToLower(fb.Char)] = true
				}
			}

			var feedbackString strings.Builder
			for i, fb := range detailedFeedback {
				colorToUse := ColorWhite
				switch fb.Status {
				case "green":
					colorToUse = ColorGreen
				case "yellow":
					colorToUse = ColorYellow
				}
				feedbackString.WriteString(colorToUse)
				feedbackString.WriteString(strings.ToUpper(string(fb.Char)))
				feedbackString.WriteString(ColorReset)
				if i < len(detailedFeedback)-1 {
					feedbackString.WriteString("")
				}
			}

			fmt.Printf("Feedback: %s\n", feedbackString.String())

			var remainingCharsDisplay []string
			for charCode := 'A'; charCode <= 'Z'; charCode++ {
				lowerChar := unicode.ToLower(charCode)
				if _, isIncorrect := guessedIncorrectLetters[lowerChar]; !isIncorrect {
					remainingCharsDisplay = append(remainingCharsDisplay, string(charCode))
				}
			}
			fmt.Printf("Remaining letters: %s \n", strings.Join(remainingCharsDisplay, " "))
			fmt.Printf("Attempts remaining:  %d\n", attemptsLeft)

			if attemptsLeft == 0 {
				fmt.Printf("Game over. The correct word was: %s\n", secretWord)
				break
			}
		}
	} else {
		eofDuringGuessLoop = true
	}

	gameConcludedNatively := gameWon || (attemptsLeft == 0)

	if eofDuringUsername || (eofDuringGuessLoop && !gameConcludedNatively) {
		// EOF interrupted game
	} else if gameConcludedNatively {
		attemptsMade := maxAttempts - attemptsLeft
		resultStatus := "loss"
		if gameWon {
			resultStatus = "win"
		}

		if username != "" {
			recordStats(username, secretWord, attemptsMade, resultStatus)
		}

		fmt.Print("Do you want to see your stats? (yes/no): ")
		if gameInputScanner.Scan() {
			seeStatsAnswer := strings.ToLower(strings.TrimSpace(gameInputScanner.Text()))
			if seeStatsAnswer == "yes" {
				if username != "" {
					displayStats(username, gameInputScanner)
				} else {
					fmt.Println("No username was entered; cannot display specific stats.")
					fmt.Print("Press Enter to exit...\n")
					if gameInputScanner.Scan() {
					}
				}
			}
		}
	}
}

func loadWords(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", filename, err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(word) == wordLength {
			isLowercaseAlpha := true
			for _, r := range word {
				if r < 'a' || r > 'z' {
					isLowercaseAlpha = false
					break
				}
			}
			if isLowercaseAlpha {
				words = append(words, word)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("no valid %d-letter words found in %s", wordLength, filename)
	}
	return words, nil
}

func isWordInList(word string, wordList []string) bool {
	for _, w := range wordList {
		if w == word {
			return true
		}
	}
	return false
}

func generateFeedback(guess, secret string) ([]LetterFeedback, bool) {
	n := len(secret)
	feedback := make([]LetterFeedback, n)
	guessRunes := []rune(guess)
	secretRunes := []rune(secret)
	secretMutable := []rune(secret)

	for i := 0; i < n; i++ {
		feedback[i] = LetterFeedback{Char: guessRunes[i], Status: "white"}
	}

	for i := 0; i < n; i++ {
		if guessRunes[i] == secretRunes[i] {
			feedback[i].Status = "green"
			secretMutable[i] = 0
		}
	}

	for i := 0; i < n; i++ {
		if feedback[i].Status == "green" {
			continue
		}

		for j := 0; j < n; j++ {
			if secretMutable[j] == guessRunes[i] {
				feedback[i].Status = "yellow"
				secretMutable[j] = 0
				break
			}
		}
	}

	isWin := true
	for _, fb := range feedback {
		if fb.Status != "green" {
			isWin = false
			break
		}
	}
	return feedback, isWin
}

func recordStats(username, secretWord string, attempts int, result string) {
	file, err := os.OpenFile(statsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{username, secretWord, strconv.Itoa(attempts), result}
	_ = writer.Write(record)
}

func displayStats(username string, scanner *bufio.Scanner) {
	file, err := os.Open(statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No stats available yet for %s.\n", username)
		} else {
			fmt.Println("Could not retrieve stats.")
		}
		fmt.Print("Press Enter to exit...\n")
		if scanner.Scan() {
		}
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	totalGamesPlayed := 0
	gamesWon := 0
	totalAttemptsAllGames := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(record) == 4 && record[0] == username {
			attempts, errAtoi := strconv.Atoi(record[2])
			if errAtoi != nil {
				continue
			}
			gameResult := record[3]

			totalGamesPlayed++
			totalAttemptsAllGames += attempts
			if gameResult == "win" {
				gamesWon++
			}
		}
	}

	fmt.Printf("Stats for %s:\n", username)
	fmt.Printf("Games played: %d\n", totalGamesPlayed)
	fmt.Printf("Games won: %d\n", gamesWon)

	avgAttempts := 0.0
	if totalGamesPlayed > 0 {
		avgAttempts = float64(totalAttemptsAllGames) / float64(totalGamesPlayed)
	}
	fmt.Printf("Average attempts per game: %.2f\n", avgAttempts)

	fmt.Print("Press Enter to exit...\n")
	if scanner.Scan() {
	}
}
