package main

import (
	"fmt"
	"gocore/utils"
	"os"

	flag "github.com/spf13/pflag"
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
		parentsFlag := mkdirCmd.BoolP("parents", "p", false, "Create any missing intermediate pathname components.")
		mkdirCmd.Parse(os.Args[2:])
		err := utils.Mkdir(*permFlag, *parentsFlag, mkdirCmd.Args())
		if err != nil {
			fmt.Println(err)
		}

	case "rm":
		rmCmd := flag.NewFlagSet("rm", flag.ExitOnError)
		interactiveFlag := rmCmd.BoolP("interactive", "i", false, "prompt before every removal")
		forceFlag := rmCmd.BoolP("force", "f", false, "Do not prompt for confirmation. Do not write diagnostic messages or modify the exit status in the case of no file operands, or in the case of operands that do not exist.")
		recursiveFlag := rmCmd.BoolP("recursive", "r", false, "Remove file hierarchies.")
		rmCmd.Parse(os.Args[2:])
		err := utils.Rm(*interactiveFlag, *forceFlag, *recursiveFlag, rmCmd.Args())
		if err != nil {
			fmt.Println(err)
		}

	case "cat":
		catCmd := flag.NewFlagSet("cat", flag.ExitOnError)
		bytesFlag := catCmd.BoolP("bytes", "u", false, "Write bytes from the input file to the standard output without delay as each is read.")
		catCmd.Parse(os.Args[2:])
		err := utils.Cat(*bytesFlag, catCmd.Args()...)
		if err != nil {
			fmt.Println(err)
		}

	case "head":
		headCmd := flag.NewFlagSet("head", flag.ExitOnError)
		linesFlag := headCmd.IntP("lines", "n", 10, "The first number lines of each input file")
		headCmd.Parse(os.Args[2:])
		err := utils.Head(*linesFlag, headCmd.Args()...)
		if err != nil {
			fmt.Println(err)
		}

	case "tail":
		tailCmd := flag.NewFlagSet("tail", flag.ExitOnError)
		bytesFlag := tailCmd.StringP("bytes", "c", "0", "output the last NUM bytes; or use -c +NUM to output starting with byte NUM of each file")
		linesFlag := tailCmd.StringP("lines", "n", "10", "output the last NUM lines, instead of the last 10; or use -n +NUM to skip NUM-1 lines at the start")
		followFlag := tailCmd.BoolP("follow", "f", false, "output appended data as the file grows;")
		tailCmd.Parse(os.Args[2:])

		err := utils.Tail(tailCmd.Arg(0), *bytesFlag, *linesFlag, *followFlag)
		if err != nil {
			fmt.Println(err)
		}

	case "cp":
		utils.Cp(flag.Arg(1), flag.Arg(2))

	case "cal":
		calCmd := flag.NewFlagSet("cal", flag.ExitOnError)
		calCmd.Parse(os.Args[2:])
		utils.Cal(calCmd.Arg(0), calCmd.Arg(1))

	case "cmp":
		cmpCmd := flag.NewFlagSet("cmp", flag.ExitOnError)
		verboseFlag := cmpCmd.BoolP("verbose", "l", false, "output byte numbers and differing byte values")
		quietFlag := cmpCmd.BoolP("quiet", "s", false, "suppress all normal output")
		cmpCmd.Parse(os.Args[2:])
		_, _, err := utils.Cmp(cmpCmd.Arg(0), cmpCmd.Arg(1), *verboseFlag, *quietFlag)
		if err != nil {
			fmt.Println(err)
		}

	case "mv":
		mvCmd := flag.NewFlagSet("mv", flag.ExitOnError)
		interactiveFlag := mvCmd.BoolP("interactive", "i", false, "prompt before overwrite")
		forceFlag := mvCmd.BoolP("force", "f", false, "do not prompt before overwriting")
		mvCmd.Parse(os.Args[2:])
		utils.Mv(mvCmd.Args(), *interactiveFlag, *forceFlag)

	case "tee":
		teeCmd := flag.NewFlagSet("tee", flag.ExitOnError)
		appendFlag := teeCmd.BoolP("append", "a", false, "append to the given FILEs, do not overwrite")
		ignoreInterruptsFlag := teeCmd.BoolP("ignore-interrupts", "i", false, "ignore interrupt signals")
		teeCmd.Parse(os.Args[2:])
		utils.Tee(os.Stdin, teeCmd.Args(), *appendFlag, *ignoreInterruptsFlag)

	case "ln":
		lnCmd := flag.NewFlagSet("ln", flag.ExitOnError)
		symlinkFlag := lnCmd.BoolP("symbolic", "s", false, "make symbolic links instead of hard links")
		forceFlag := lnCmd.BoolP("force", "f", false, "remove existing destination files")
		logicalFlag := lnCmd.BoolP("logical", "L", false, "dereference TARGETs that are symbolic links")
		physicalFlag := lnCmd.BoolP("physical", "P", false, "make hard links directly to symbolic links")
		lnCmd.Parse(os.Args[2:])
		utils.Ln(lnCmd.Args(), *symlinkFlag, *forceFlag, *logicalFlag, *physicalFlag)
	case "comm":
		commCmd := flag.NewFlagSet("comm", flag.ExitOnError)
		com1Flag := commCmd.BoolP("1", "1", false, "suppress column 1 (lines unique to FILE1)")
		com2Flag := commCmd.BoolP("2", "2", false, "suppress column 2 (lines unique to FILE2)")
		com3Flag := commCmd.BoolP("3", "3", false, "suppress column 3 (lines that appear in both files)")
		commCmd.Parse(os.Args[2:])
		utils.Comm(commCmd.Arg(0), commCmd.Arg(1), *com1Flag, *com2Flag, *com3Flag)

	case "chown":
		chownCmd := flag.NewFlagSet("chown", flag.ExitOnError)
		noDereferenceFlag := chownCmd.BoolP("no-dereference", "d", false, "affect symbolic links instead of any referenced file (useful only on systems that can change the ownership of a symlink)")
		recursiveFlag := chownCmd.BoolP("reccursive", "R", false, "operate on files and directories recursively")
		physicalFlag := chownCmd.BoolP("physical", "P", false, "do not traverse any symbolic links")
		logicalFlag := chownCmd.BoolP("logical", "L", false, "traverse every symbolic link to a directory encountered")
		hybridFlag := chownCmd.BoolP("Hybrid", "H", false, "if a command line argument is a symbolic link to a directory, traverse it")
		chownCmd.Parse(os.Args[2:])
		utils.Chown(chownCmd.Arg(0), chownCmd.Args()[1:], *noDereferenceFlag, *recursiveFlag, *physicalFlag, *logicalFlag, *hybridFlag)

	case "touch":
		touchCmd := flag.NewFlagSet("touch", flag.ExitOnError)
		noCreateFlag := touchCmd.BoolP("no-create", "c", false, "do not create any files")
		accessOnlyFlag := touchCmd.BoolP("access", "a", false, "change only the access time")
		modifyOnlyFlag := touchCmd.BoolP("modify", "m", false, "change only the modification time")
		dateFlag := touchCmd.StringP("date", "d", "", "parse STRING and use it instead of current time")
		timestampFlag := touchCmd.StringP("timestamp", "t", "", "use specified time instead of current time, with a date-time format that differs from -d's")
		referenceFlag := touchCmd.BoolP("reference", "r", false, "use this file's times instead of current time")
		touchCmd.Parse(os.Args[2:])
		err := utils.Touch(touchCmd.Args(), *noCreateFlag, *accessOnlyFlag, *modifyOnlyFlag, *dateFlag, *timestampFlag, *referenceFlag)
		if err != nil {
			fmt.Println(err)
		}
	}

}
