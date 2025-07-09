package utils

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type Flags struct {
	Name string
	Flag []string
	Desc []string
}

// String interface for flags
func (f Flags) String() string {
	var flag string
	s := f.Name + "\n\tAvailable flags:\n"
	for i := range len(f.Flag) {
		flag += fmt.Sprintf("\t\t%s\t%s\n", f.Flag[i], f.Desc[i])
	}
	return s + flag
}

// make ls column view
func columnise(w *tabwriter.Writer, opt []string) {
	for i := 0; i < len(opt); i += 3 {
		if i == len(opt)-1 {
			fmt.Fprintln(w, fmt.Sprintf("%s\t", opt[i]))
		} else if i == len(opt)-2 {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t", opt[i], opt[i+1]))
		} else {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s", opt[i], opt[i+1], opt[i+2]))
		}
	}
}

// append indicator (one of /*@|) to entries
func classifyVer(dir string, options *[]string) {
	err := os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for idx, file := range *options {
		fi, err := os.Lstat(file)
		if err != nil {
			log.Fatal(err)
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			(*options)[idx] = file + "/"
		case mode&fs.ModeSymlink != 0:
			(*options)[idx] = file + "@"
		case mode&fs.ModeNamedPipe != 0:
			(*options)[idx] = file + "|"
		}
	}
}

func Ls(dir string, allDir bool, column bool, classify bool, help bool) {
	flag := Flags{
		Name: "ls: - list directory contents",
		Flag: []string{"-A , --almost-all", "-F, --classify", "-C"},
		Desc: []string{"Write out all directory entries, including those whose names begin with a <period> ( '.' ) but excluding the entries dot and dot-dot (if they exist).", "This flag appends a character to the end of each filename to indicate its type.", "list entries by columns"},
	}

	// prints help (-h --help)
	if help {
		fmt.Println(flag)
		return
	}

	// if no dir is specified use the current dir
	if dir == "" {
		dir, _ = os.Getwd()
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var result []string

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") && allDir {
			result = append(result, file.Name())
		} else {
			result = append(result, file.Name())
		}

	}

	if classify {
		classifyVer(dir, &result)
	}

	// checks if column view is true and print the result in column view
	if column {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 0, 10, ' ', 0)

		sort.Strings(result)
		columnise(w, result)

		w.Flush()
		return
	}

}

func Mkdir(perm int, dir []string, help bool) {
	if help || len(dir) < 1 {
		fmt.Println("mkdir - make directories")
		fmt.Println("\tAvailable flags:")
		fmt.Printf("\t\t-m, --mode=MODE\tset file mode (e.g -m 755)\n")
		return

	}
	for _, files := range dir {
		err := os.Mkdir(files, os.FileMode(perm))
		if err != nil && os.IsExist(err) {
			log.Printf("Directory '%s' already exists.", files)
		} else if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
	}

}

func Rm(dir []string, help bool) {
	if help || len(dir) < 1 {
		fmt.Println("rm - remove files or directories")
		fmt.Println("\tAvailable flags:")
		fmt.Printf("\t\t-m, --mode=MODE\tset file mode (e.g -m 755)\n")
		return
	}
	for _, files := range dir {
		err := os.Remove(files)
		if err != nil && os.IsExist(err) {
			log.Printf("Directory '%s' is not empty.", files)
		} else if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
	}
}

func Cat(file string, help bool) {
	if help || len(file) < 1 {
		fmt.Println("cat - print on the standard output")
		fmt.Println("\tAvailable flags:")
		fmt.Printf("\t\t-m, --mode=MODE\tset file mode (e.g -m 755)\n")
		return
	}
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)
}

func Head(file string, help bool) {
	if help || len(file) < 1 {
		fmt.Println("Head - output the first part of files")
		fmt.Println("\tAvailable flags:")
		fmt.Printf("\t\t-m, --mode=MODE\tset file mode (e.g -m 755)\n")
		return
	}
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := 0

	for scanner.Scan() && lines < 10 {
		fmt.Println(scanner.Text())
		lines++
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func Tail(file string, help bool, bytes int) {
	flag := Flags{
		Name: "Tail - output the last part of files",
		Flag: []string{"-c , --bytes=NUM", "-t, --test=00"},
		Desc: []string{"output the last NUM bytes;", "IDK BRO"},
	}
	if help || len(file) < 1 {
		fmt.Println(flag)
		return
	}

	if bytes != 0 {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		buf := make([]byte, bytes)
		stat, statErr := f.Stat()
		if statErr != nil {
			panic(statErr)
		}
		start := stat.Size() - int64(bytes)
		_, err = f.ReadAt(buf, start)
		if err == nil {
			fmt.Printf("%s\n", buf)
		}
	}

}

func Cp(src string, dst string, help bool) {
	flag := Flags{
		Name: "Cp - copy files and directories",
		Flag: []string{"-c , --bytes=NUM", "-t, --test=00"},
		Desc: []string{"output the last NUM bytes;", "IDK BRO"},
	}
	if help || len(src) < 1 {
		fmt.Println(flag)
		return
	}

	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(dst, data, 0755)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
}

func Cal(month string, year string) {

	monthInt, err := strconv.Atoi(month)
	if err != nil {
		monthInt = int(time.Now().Month())
	}

	yearInt, _ := strconv.Atoi(year)
	if err != nil {
		yearInt = time.Now().Year()
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	currentTime := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.Local)
	lastDayOfMonth := time.Date(currentTime.Year(), currentTime.Month()+1, 0, 0, 0, 0, 0, currentTime.Location())
	firstOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())

	fmt.Fprintf(w, "\t\t%s %d\n", currentTime.Month(), currentTime.Year())

	fmt.Fprintln(w, "Su\tMo\tTu\tWe\tTh\tFr\tSa\t")

	for range int(firstOfMonth.Weekday()) {
		fmt.Fprintf(w, " \t")
	}

	for i := 1; i <= int(lastDayOfMonth.Day()); i++ {
		fmt.Fprintf(w, "%d\t", i)

		if (i+int(firstOfMonth.Weekday()))%7 == 0 {
			fmt.Fprintln(w)
		}

	}
	fmt.Fprintln(w)

	w.Flush()

}
