package head

import (
	"bufio"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
)

func head(lines int, files ...string) error {
	if lines <= 0 {
		return fmt.Errorf("Error: number of lines is less than 1")
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %w", file, err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineCount := 0

		if len(files) > 1 {
			fmt.Printf("\n\n==> %s <==\n\n", file)
		}
		for scanner.Scan() && lineCount < lines {
			fmt.Println(scanner.Text())
			lineCount++
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading standard input: %w", err)
		}

	}

	return nil
}

func Exec() {
	headCmd := flag.NewFlagSet("head", flag.ExitOnError)
	linesFlag := headCmd.IntP("lines", "n", 10, "The first number lines of each input file")
	headCmd.Parse(os.Args[2:])
	err := head(*linesFlag, headCmd.Args()...)
	if err != nil {
		fmt.Println(err)
	}

}
