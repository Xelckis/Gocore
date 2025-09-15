package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func MvPrompt(file string) bool {
	var answer string
	fmt.Printf("mv: Overwrite '%s'? ", file)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer = strings.ToLower(scanner.Text())
	}

	if !(answer == "y") && !(answer == "yes") {
		return false
	}
	return true
}

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err == nil && !info.IsDir() {
		return true
	}
	return false
}
