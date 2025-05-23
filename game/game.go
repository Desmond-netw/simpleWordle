package game

import (
	"bufio"
	"fmt"
	"strings"

	"simpleWordle/io"

	"simpleWordle/model"
)

const maxAttempts = 6

var allLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Play(scanner *bufio.Scanner, user *model.User, secret string) {
	fmt.Println("Welcome to Wordle! Guess the 5-letter word.")
	secret = strings.ToUpper(secret)
	attempts := 0
	guessed := false
	remainingLetters := make(map[rune]bool)
	for _, r := range allLetters {
		remainingLetters[r] = true
	}

	for attempts < maxAttempts {
		fmt.Print("Enter your guess: ")
		if !scanner.Scan() {
			break
		}
		guess := strings.ToUpper(scanner.Text())
		if len(guess) != 5 {
			fmt.Println("Guess must be 5 letters.")
			continue
		}

		attempts++
		feedback := getFeedback(secret, guess)
		fmt.Println("Feedback:", feedback)

		for _, r := range guess {
			delete(remainingLetters, r)
		}
		printRemainingLetters(remainingLetters)
		fmt.Println("Attempts remaining:", maxAttempts-attempts)

		if guess == secret {
			guessed = true
			break
		}
	}

	user.RecordGame(guessed, attempts)

	if !guessed {
		fmt.Println("Out of attempts! The word was:", secret)
	}

	io.AppendStats("stats.csv", []string{user.Name, secret, fmt.Sprint(attempts), map[bool]string{true: "win", false: "loss"}[guessed]})

	showStatsPrompt(scanner, user)
}

func getFeedback(secret, guess string) string {
	feedback := ""
	used := make([]bool, 5)

	for i := 0; i < 5; i++ {
		if guess[i] == secret[i] {
			feedback += "\u001B[32m" + string(guess[i]) + "\u001B[0m" // green
			used[i] = true
		} else {
			feedback += "_"
		}
	}

	for i := 0; i < 5; i++ {
		if feedback[i] == '_' {
			found := false
			for j := 0; j < 5; j++ {
				if !used[j] && guess[i] == secret[j] {
					found = true
					used[j] = true
					break
				}
			}
			if found {
				feedback = feedback[:i] + "\u001B[33m" + string(guess[i]) + "\u001B[0m" + feedback[i+1:]
			} else {
				feedback = feedback[:i] + "\u001B[37m" + string(guess[i]) + "\u001B[0m" + feedback[i+1:]
			}
		}
	}

	return feedback
}

func printRemainingLetters(letters map[rune]bool) {
	fmt.Print("Remaining letters: ")
	for _, r := range allLetters {
		if letters[r] {
			fmt.Printf("%c ", r)
		}
	}
	fmt.Println()
}

func showStatsPrompt(scanner *bufio.Scanner, user *model.User) {
	fmt.Print("Do you want to see your stats? (yes/no): ")
	if !scanner.Scan() {
		return
	}
	if strings.ToLower(scanner.Text()) == "yes" {
		fmt.Println("Stats for", user.Name+":")
		g, w, avg := user.Stats()
		fmt.Println("Games played:", g)
		fmt.Println("Games won:", w)
		fmt.Printf("Average attempts per game: %.2f\n", avg)
		fmt.Println("Press Enter to exit...")
		scanner.Scan()
	}
}
