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
		AlmostallDir := lsCmd.BoolP("almost-all", "A", false, "Lists all entries, including hidden files (those starting with a .), but excludes the current directory (.) and the parent directory (..).")
		columnFlag := lsCmd.BoolP("column", "C", false, "Forces the output into multiple columns")
		classifyFlag := lsCmd.BoolP("classify", "F", false, "This flag appends a character to the end of each filename to indicate its type (/*@|).")
		recursiveFlag := lsCmd.BoolP("recursive", "R", false, "list subdirectories recursively")
		allDir := lsCmd.BoolP("all", "a", false, "do not ignore entries starting with .")
		longListingFlag := lsCmd.BoolP("long-listing", "l", false, "use a long listing format")
		sortSizeFlag := lsCmd.BoolP("size", "S", false, "sort by file size, largest first")
		sizeKbFlag := lsCmd.BoolP("kibibytes", "k", false, "default to 1024-byte blocks for file system usage; used only with -s and per directory totals")
		omitOwnerFlag := lsCmd.BoolP("omit-owner", "g", false, "like -l, but do not list owner")
		omitGroupFlag := lsCmd.BoolP("omit-group", "o", false, "like -l, but do not list group information")
		changeTimeFlag := lsCmd.BoolP("change-time", "c", false, "Use time of last modification of the file status information")
		streamFormatFlag := lsCmd.BoolP("stream-format", "m", false, "fill width with a comma separated list of entries")
		numericUidGidFlag := lsCmd.BoolP("numeric-uid-gid", "n", false, "Turn on the −l (ell) option, but when writing the file’s owner or group, write the file’s numeric UID or GID rather than the user or group name.")
		showInodeFlag := lsCmd.BoolP("show-inode", "i", false, "For each file, write the file’s file serial number (inode)")
		dereferenceFlag := lsCmd.BoolP("dereference", "L", false, "when showing file information for a symbolic link, show information for the file the link references rather than for the link itself")
		onePerLineFlag := lsCmd.BoolP("one-per-line", "1", false, "list one file per line")
		sortByMtimeFlag := lsCmd.BoolP("sort-mtime", "t", false, "sort by time, newest first")
		indicatorStyleFlag := lsCmd.BoolP("indicator-style", "p", false, "append / indicator to directories")
		hideControlCharsFlag := lsCmd.BoolP("hide-control-chars", "q", false, "print ? instead of nongraphic characters")
		reverseSortFlag := lsCmd.BoolP("reverse-sort", "r", false, "reverse order while sorting")
		accessTimeFlag := lsCmd.BoolP("access-time", "u", false, "Use time of last access instead of last modification of the file for sorting (−t) or writing (−l).")
		noSortFlag := lsCmd.BoolP("no-sort", "f", false, "do not sort")
		lsCmd.Parse(os.Args[2:])
		err := utils.Ls(lsCmd.Args(), *AlmostallDir, *columnFlag, *classifyFlag, *recursiveFlag, *allDir, *longListingFlag, *sortSizeFlag, *sizeKbFlag, *streamFormatFlag, *omitOwnerFlag, *omitGroupFlag, *changeTimeFlag, *numericUidGidFlag, *showInodeFlag, *dereferenceFlag, *onePerLineFlag, *sortByMtimeFlag, *indicatorStyleFlag, *hideControlCharsFlag, *reverseSortFlag, *accessTimeFlag, *noSortFlag)

		if err != nil {
			fmt.Println(err)
		}

	case "mkdir":
		mkdirCmd := flag.NewFlagSet("mkdir", flag.ExitOnError)
		permFlag := mkdirCmd.IntP("mode", "m", 0755, "set file mode")
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
		cpCmd := flag.NewFlagSet("cp", flag.ExitOnError)
		followSymbolicFlag := cpCmd.BoolP("follow-symbolic", "H", false, "follow command-line symbolic links in SOURCE")
		recursiveFlag := cpCmd.BoolP("recursive", "r", false, "copy directories recursively")
		dereferenceFlag := cpCmd.BoolP("dereference", "L", false, "always follow symbolic links in SOURCE")
		noDereferenceFlag := cpCmd.BoolP("no-dereference", "P", false, "never follow symbolic links in SOURCE")
		preserveFlag := cpCmd.BoolP("preserve", "p", false, "preserve the file attributes")
		cpCmd.Parse(os.Args[2:])
		err := utils.Cp(cpCmd.Args(), *followSymbolicFlag, *recursiveFlag, *dereferenceFlag, *noDereferenceFlag, *preserveFlag)
		if err != nil {
			fmt.Println(err)
		}

	case "cal":
		calCmd := flag.NewFlagSet("cal", flag.ExitOnError)
		calCmd.Parse(os.Args[2:])
		utils.Cal(calCmd.Args())

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
		err := utils.Mv(mvCmd.Args(), *interactiveFlag, *forceFlag)

		if err != nil {
			fmt.Println(err)
		}

	case "tee":
		teeCmd := flag.NewFlagSet("tee", flag.ExitOnError)
		appendFlag := teeCmd.BoolP("append", "a", false, "append to the given FILEs, do not overwrite")
		ignoreInterruptsFlag := teeCmd.BoolP("ignore-interrupts", "i", false, "ignore interrupt signals")
		teeCmd.Parse(os.Args[2:])
		err := utils.Tee(os.Stdin, teeCmd.Args(), *appendFlag, *ignoreInterruptsFlag)

		if err != nil {
			fmt.Println(err)
		}

	case "ln":
		lnCmd := flag.NewFlagSet("ln", flag.ExitOnError)
		symlinkFlag := lnCmd.BoolP("symbolic", "s", false, "make symbolic links instead of hard links")
		forceFlag := lnCmd.BoolP("force", "f", false, "remove existing destination files")
		logicalFlag := lnCmd.BoolP("logical", "L", false, "dereference TARGETs that are symbolic links")
		physicalFlag := lnCmd.BoolP("physical", "P", false, "make hard links directly to symbolic links")
		lnCmd.Parse(os.Args[2:])
		err := utils.Ln(lnCmd.Args(), *symlinkFlag, *forceFlag, *logicalFlag, *physicalFlag)

		if err != nil {
			fmt.Println(err)
		}
	case "comm":
		commCmd := flag.NewFlagSet("comm", flag.ExitOnError)
		com1Flag := commCmd.BoolP("1", "1", false, "suppress column 1 (lines unique to FILE1)")
		com2Flag := commCmd.BoolP("2", "2", false, "suppress column 2 (lines unique to FILE2)")
		com3Flag := commCmd.BoolP("3", "3", false, "suppress column 3 (lines that appear in both files)")
		commCmd.Parse(os.Args[2:])
		err := utils.Comm(commCmd.Arg(0), commCmd.Arg(1), *com1Flag, *com2Flag, *com3Flag)

		if err != nil {
			fmt.Println(err)
		}

	case "chown":
		chownCmd := flag.NewFlagSet("chown", flag.ExitOnError)
		noDereferenceFlag := chownCmd.BoolP("no-dereference", "d", false, "affect symbolic links instead of any referenced file (useful only on systems that can change the ownership of a symlink)")
		recursiveFlag := chownCmd.BoolP("reccursive", "R", false, "operate on files and directories recursively")
		physicalFlag := chownCmd.BoolP("physical", "P", false, "do not traverse any symbolic links")
		logicalFlag := chownCmd.BoolP("logical", "L", false, "traverse every symbolic link to a directory encountered")
		hybridFlag := chownCmd.BoolP("Hybrid", "H", false, "if a command line argument is a symbolic link to a directory, traverse it")
		chownCmd.Parse(os.Args[2:])
		err := utils.Chown(chownCmd.Arg(0), chownCmd.Args()[1:], *noDereferenceFlag, *recursiveFlag, *physicalFlag, *logicalFlag, *hybridFlag)

		if err != nil {
			fmt.Println(err)
		}
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

	case "uniq":
		uniqCmd := flag.NewFlagSet("uniq", flag.ExitOnError)
		counterFlag := uniqCmd.BoolP("count", "c", false, "prefix lines by the number of occurrences")
		repeatedFlag := uniqCmd.BoolP("repeated", "d", false, "only print duplicate lines, one for each group")
		uniqueFlag := uniqCmd.BoolP("unique", "u", false, "only print unique lines")
		fieldsFlag := uniqCmd.UintP("skip-fields", "f", 1, "avoid comparing the first N fields")
		charsFlag := uniqCmd.UintP("skip-chars", "s", 1, "avoid comparing the first N characters")
		uniqCmd.Parse(os.Args[2:])
		err := utils.Uniq(uniqCmd.Arg(0), uniqCmd.Arg(1), *repeatedFlag, *uniqueFlag, *counterFlag, *fieldsFlag, *charsFlag)

		if err != nil {
			fmt.Println(err)
		}
	case "cut":
		cutCmd := flag.NewFlagSet("cut", flag.ExitOnError)
		charFlag := cutCmd.StringP("characters", "c", "", "select only these characters")
		fieldFlag := cutCmd.StringP("fields", "f", "", "select only these fields;  also print any line that contains no delimiter character, unless the -s option is specified")
		delimiterFlag := cutCmd.StringP("delimiter", "d", "", "use DELIM instead of TAB for field delimiter")
		onlyDelimited := cutCmd.BoolP("only-delimited", "s", false, "do not print lines not containing delimiters")
		cutCmd.Parse(os.Args[2:])
		err := utils.Cut(cutCmd.Args(), *charFlag, *fieldFlag, *delimiterFlag, *onlyDelimited)

		if err != nil {
			fmt.Println(err)
		}

	case "more":
		moreCmd := flag.NewFlagSet("more", flag.ExitOnError)
		clearFlag := moreCmd.BoolP("clean-print", "c", false, "Do not scroll. Instead, paint each screen from the top, clearing the remainder of each line as it is displayed.")
		linesFlag := moreCmd.IntP("lines", "n", 0, "Specify the number of lines per screenful. The number argument is a positive decimal integer. The --lines option shall override any values obtained from any other source, such as number of lines reported by terminal.")
		caseFlag := moreCmd.BoolP("case-insensitive", "i", false, "Perform pattern matching in searches without regard to case")
		commandFlag := moreCmd.StringP("command", "p", "", "Each time a screen from a new file is displayed or redisplayed (including as a result of more commands; for example, :p), execute the more command(s) in the command arguments in the order specified, as if entered by the user after the first screen has been displayed.")
		tagFlag := moreCmd.StringP("tag", "t", "", "Start displaying the file from the first line containing the specified tag. If the tag is not found, display begins from the start of the file.")
		squeezeFlag := moreCmd.BoolP("squeeze", "s", false, "Squeeze multiple blank lines into one.")
		moreCmd.Parse(os.Args[2:])
		err := utils.More(moreCmd.Args(), *clearFlag, *caseFlag, *squeezeFlag, *linesFlag, *commandFlag, *tagFlag)

		if err != nil {
			fmt.Println(err)
		}
	}

}
