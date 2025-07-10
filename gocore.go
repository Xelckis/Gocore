package main

import (
	flag "github.com/spf13/pflag"
	"gocore/utils"
	"os"
)

func main() {

	switch os.Args[1] {
	case "ls":
		lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
		allDir := lsCmd.BoolP("almost-all", "A", false, "Lists all entries, including hidden files (those starting with a .), but excludes the current directory (.) and the parent directory (..).")
		columnFlag := lsCmd.BoolP("column", "C", false, "Forces the output into multiple columns")
		classifyFlag := lsCmd.BoolP("classify", "F", false, "This flag appends a character to the end of each filename to indicate its type (/*@|).")

		lsCmd.Parse(os.Args[2:])
		utils.Ls(lsCmd.Arg(0), *allDir, *columnFlag, *classifyFlag)
	case "mkdir":
		mkdirCmd := flag.NewFlagSet("mkdir", flag.ExitOnError)
		permFlag := mkdirCmd.IntP("mode", "m", 0755, "set file mode (Default: 0755)")
		mkdirCmd.Parse(os.Args[2:])
		utils.Mkdir(*permFlag, mkdirCmd.Args())
	case "rm":
		args := flag.Args()
		utils.Rm(args[1:])
	case "cat":
		catCmd := flag.NewFlagSet("cat", flag.ExitOnError)
		bytesFlag := catCmd.BoolP("bytes", "u", false, "Write bytes from the input file to the standard output without delay as each is read.")
		catCmd.Parse(os.Args[2:])
		utils.Cat(catCmd.Arg(0), *bytesFlag)
	case "head":
		utils.Head(flag.Arg(1))
	case "tail":
		utils.Tail(flag.Arg(1), 0)
	case "cp":
		utils.Cp(flag.Arg(1), flag.Arg(2))
	case "cal":
		utils.Cal(flag.Arg(1), flag.Arg(2))
	}

}
