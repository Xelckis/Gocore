package cat

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"os"
)

func cat(printByteByByte bool, files ...string) error {

	for _, file := range files {
		if printByteByByte {
			catBytePrinter(file)
			return nil
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("cannot read the file %s: %w", file, err)
		}
		os.Stdout.Write(data)
	}
	return nil
}

func catBytePrinter(file string) error {
	files, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': %w", file, err)

	}
	defer files.Close()

	b := make([]byte, 1)

	for {
		n, err := files.Read(b)

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("Error reading byte: %w", err)
		}

		if n > 0 {
			fmt.Printf("%c", b[0])

		}
	}
	return nil

}

func Exec() {
	catCmd := flag.NewFlagSet("cat", flag.ExitOnError)
	bytesFlag := catCmd.BoolP("bytes", "u", false, "Write bytes from the input file to the standard output without delay as each is read.")
	catCmd.Parse(os.Args[2:])
	err := cat(*bytesFlag, catCmd.Args()...)
	if err != nil {
		fmt.Println(err)
	}

}
