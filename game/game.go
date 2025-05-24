package game

import (
	"bufio"
	"fmt"
	"io" // Standard io package for io.EOF
	"strings"
	"unicode" // For unicode.ToLower

	kwio "simpleWordle/io" // Alias for your custom koodWordle/io
)

// ANSI color codes
const (
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Gray   = "\033[37m"
	Reset  = "\033[0m"
)

const (
	MaxAttempts = 6
	WordLength  = 5
)

func isAllLowercaseLetters(s string) bool {
	for _, ch := range s {
		if ch < 'a' || ch > 'z' {
			return false
		}
	}
	return true
}

// getFeedback generates the colored feedback string.
func getFeedback(guess, target string) string {
	feedbackChars := make([]string, WordLength)
	targetRunes := []rune(target)
	guessRunes := []rune(guess)

	// Create a mutable frequency map of target letters for accurate handling of duplicates.
	targetFreq := make(map[rune]int)
	for _, r := range targetRunes {
		targetFreq[r]++
	}

	// First pass: identify correct positions (Green)
	for i := 0; i < WordLength; i++ {
		if guessRunes[i] == targetRunes[i] {
			// Convert to uppercase for display
			feedbackChars[i] = Green + strings.ToUpper(string(guessRunes[i])) + Reset
			targetFreq[guessRunes[i]]-- // Decrement count for used letter in target
		}
	}

	// Second pass: identify correct letters in wrong positions (Yellow) or not in word (Gray)
	for i := 0; i < WordLength; i++ {
		if feedbackChars[i] != "" { // Already marked green in first pass
			continue
		}

		if targetFreq[guessRunes[i]] > 0 { // Letter exists in target and hasn't been used for a green match
			// Convert to uppercase for display
			feedbackChars[i] = Yellow + strings.ToUpper(string(guessRunes[i])) + Reset
			targetFreq[guessRunes[i]]-- // Decrement count for used letter
		} else {
			// Convert to uppercase for display
			feedbackChars[i] = Gray + strings.ToUpper(string(guessRunes[i])) + Reset
		}
	}

	return strings.Join(feedbackChars, "") // No spaces between colored letters
}

// Play function now accepts bufio.Reader
func Play(reader *bufio.Reader, username, targetWord string) {
	fmt.Println("Welcome to Wordle! Guess the 5-letter word.")
	attempts := 0
	// usedLetters map will store letters that are NOT in the target word
	usedLetters := make(map[rune]bool)

	for attempts < MaxAttempts {
		fmt.Print("Enter your guess:  ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("Error reading input:", err)
			return
		}
		guess := strings.TrimSpace(input)

		// Input Validation
		if !isAllLowercaseLetters(guess) {
			fmt.Print("Your guess must only contain lowercase letters.\n")
			continue
		}
		if len(guess) != WordLength {
			fmt.Print("Your guess must be exactly 5 letters long.\n")
			continue
		}
		if !kwio.IsWordValid(guess) {
			fmt.Print("Word not in list. Please enter a valid word.\n")
			continue
		}

		attempts++

		// Generate and print feedback
		feedback := getFeedback(guess, targetWord)
		fmt.Printf("Feedback: %s\n", feedback)

		// Update usedLetters based on guess and targetWord for the "Remaining letters" display.
		// Only mark letters as 'used' if they are NOT in the target word.
		for _, ch := range guess {
			if !strings.ContainsRune(targetWord, ch) {
				usedLetters[ch] = true
			}
		}

		// Display remaining letters
		// Changed from "%s \n" to "%s\n" to remove the extra trailing space
		fmt.Printf("Remaining letters: %s\n", getRemainingLetters(usedLetters))

		// Display attempts remaining
		fmt.Printf("Attempts remaining:  %d\n", MaxAttempts-attempts)

		// Check for win condition
		if guess == targetWord {
			fmt.Println("Congratulations! You've guessed the word correctly.")
			_ = kwio.SaveStats(username, targetWord, attempts, "true")
			// The expected output for winning directly proceeds to stats prompt or exit,
			// without further game loop output. So, we return here.
			return // ENSURE IMMEDIATE RETURN ON WIN
		}
	}

	// Game lost (loop finished without a win)
	fmt.Printf("Sorry, you did not guess the word. The word was: %s\n", targetWord)
	_ = kwio.SaveStats(username, targetWord, attempts, "false")
	// No stats prompt/display as per previous compilation fix
}

// Helper function to get the string of remaining letters.
func getRemainingLetters(usedLetters map[rune]bool) string {
	var sb strings.Builder
	for r := 'A'; r <= 'Z'; r++ {
		// Convert to lowercase to check against the map which stores lowercase used letters
		if !usedLetters[unicode.ToLower(r)] {
			sb.WriteRune(r)
			sb.WriteRune(' ') // Add space between letters
		}
	}
	return sb.String() // Returns string with a trailing space if letters are present, no TrimSpace.
}
