package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"simpleWordle/game"
	"simpleWordle/io"
	"simpleWordle/model"
)

func main() {
	fmt.Print("Enter your username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	wordIndex := 0
	if len(os.Args) < 2 {
		fmt.Println("Missing word index. Usage: go run . <wordIndex>")
		return
	}

	idx, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid word index.")
		return
	}
	wordIndex = idx

	words, err := io.ReadWordList("wordle-words.txt")
	if err != nil {
		fmt.Println("Error loading word list:", err)
		return
	}

	if wordIndex < 0 || wordIndex >= len(words) {
		fmt.Println("Word index out of range.")
		return
	}

	secret := words[wordIndex]
	user := model.NewUser(username)
	game.Play(scanner, user, secret)
}
