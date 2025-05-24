package io

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrEOF       = errors.New("EOF")
	wordList     []string
	wordListPath = "wordle-words.txt"
)

type GameStats struct {
	Username string
	Word     string
	Attempts int
	Result   string
}

func LoadWords() ([]string, error) {
	if wordList != nil {
		return wordList, nil
	}

	file, err := os.Open(wordListPath)
	if err != nil {
		return nil, fmt.Errorf("error opening word list: %w", err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if len(word) == 5 {
			words = append(words, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading word list: %w", err)
	}

	if len(words) == 0 {
		return nil, errors.New("word list is empty")
	}

	wordList = words
	return wordList, nil
}

func IsWordValid(word string) bool {
	words, err := LoadWords()
	if err != nil {
		return false
	}
	word = strings.ToUpper(word)
	for _, w := range words {
		if w == word {
			return true
		}
	}
	return false
}

func ReadInput() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", ErrEOF
}

func SaveStats(username, word string, attempts int, result string) error {
	file, err := os.OpenFile("stats.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening stats file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		username,
		word,
		strconv.Itoa(attempts),
		result,
	}

	return writer.Write(record)
}

func LoadUserStats(username string) ([]GameStats, error) {
	file, err := os.Open("stats.csv")
	if os.IsNotExist(err) {
		return []GameStats{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error opening stats file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading stats: %w", err)
	}

	var stats []GameStats
	for _, record := range records {
		if len(record) != 4 {
			continue
		}
		attempts, _ := strconv.Atoi(record[2])
		stats = append(stats, GameStats{
			Username: record[0],
			Word:     record[1],
			Attempts: attempts,
			Result:   record[3],
		})
	}

	var userStats []GameStats
	for _, stat := range stats {
		if stat.Username == username {
			userStats = append(userStats, stat)
		}
	}

	return userStats, nil
}
