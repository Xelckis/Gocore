package main

import (
	"fmt"

	flag "github.com/spf13/pflag"
	"gocore/utils"
)

func main() {

	allDir := flag.BoolP("almost-all", "A", false, "")
	permFlag := flag.IntP("mode", "m", 0755, "")
	helpFlag := flag.BoolP("help", "h", false, "")
	bytesFlag := flag.IntP("bytes", "c", 0, "")
	classifyFlag := flag.BoolP("classify", "F", false, "")
	columFlag := flag.BoolP("C", "C", false, "")
	flag.Parse()

	if len(flag.Args()) < 1 || (len(flag.Args()) < 1 && *helpFlag) {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of coreutils-go:\n")
		fmt.Fprintln(w, "\nAvailable flags:")
		fmt.Fprintf(w, "\tls: - list directory contents\n\t\tls [FILE]... [OPTION]...\n\n\tmkdir - make directories\n\t\tmkdir [OPTION]... DIRECTORY...")
		fmt.Fprintln(w, "\nFor more information, visit https://github.com/Xelckis/Gocore")
		return
	}

	switch flag.Arg(0) {
	case "ls":
		utils.Ls(flag.Arg(1), *allDir, *columFlag, *classifyFlag, *helpFlag)
	case "mkdir":
		args := flag.Args()
		utils.Mkdir(*permFlag, args[1:], *helpFlag)
	case "rm":
		args := flag.Args()
		utils.Rm(args[1:], *helpFlag)
	case "cat":
		utils.Cat(flag.Arg(1), *helpFlag)
	case "head":
		utils.Head(flag.Arg(1), *helpFlag)
	case "tail":
		utils.Tail(flag.Arg(1), *helpFlag, *bytesFlag)
	case "cp":
		utils.Cp(flag.Arg(1), flag.Arg(2), *helpFlag)
	case "cal":
		utils.Cal(flag.Arg(1), flag.Arg(2))
	}

}
