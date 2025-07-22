package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	osUser "os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
)

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

func catBytePrinter(file string) error {
	files, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error opening file: %w", err)

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

func tailBytePrinter(f *os.File, bytesString string) error {
	bytes, err := strconv.Atoi(strings.TrimPrefix(bytesString, "+"))
	if err != nil {
		return fmt.Errorf("cannot convert %s to int: %w", bytesString, err)
	}
	var buf []byte
	var start int64

	stat, statErr := f.Stat()
	if statErr != nil {
		return fmt.Errorf("cannot get file statistics: %w", statErr)
	}

	if strings.HasPrefix(bytesString, "+") {
		buf = make([]byte, stat.Size()-int64(bytes))
		start = int64(bytes)
	} else {
		buf = make([]byte, bytes)
		start = stat.Size() - int64(bytes)
	}

	_, err = f.ReadAt(buf, start)
	if err == nil {
		fmt.Printf("%s\n", buf)
	}
	return nil
}

func tailFollow(f *os.File, linesString string, bytesString string) error {
	if bytesString != "0" {
		tailBytePrinter(f, bytesString)
	} else {
		tailLinePrinter(f, linesString)
	}

	stat, statErr := f.Stat()
	if statErr != nil {
		return fmt.Errorf("cannot get file statistics: %w", statErr)

	}

	oldSize := stat.Size()

	for {
		time.Sleep(500 * time.Millisecond)

		stat, statErr := f.Stat()
		if statErr != nil {
			return fmt.Errorf("cannot get file statistics: %w", statErr)

		}

		newSize := stat.Size()

		if newSize > oldSize {
			f.Seek(oldSize, io.SeekStart)

			if _, err := io.Copy(os.Stdout, f); err != nil {
				return fmt.Errorf("cannot print new bytes: %w", err)
			}

			oldSize = newSize
		} else if newSize < oldSize {
			oldSize = newSize
		}
	}

}

func tailLinePrinter(f *os.File, linesString string) error {
	lines, err := strconv.Atoi(strings.TrimPrefix(linesString, "+"))
	if err != nil {
		return fmt.Errorf("cannot convert %s to int: %w", linesString, err)
	}

	scanner := bufio.NewScanner(f)

	if strings.HasPrefix(linesString, "+") {
		for i := 0; scanner.Scan(); i++ {
			if i >= lines-1 {
				println(scanner.Text())
			}
		}
	} else {
		var currentIndex = 0
		var itemsAdded = 0

		line := make([]string, lines)

		for scanner.Scan() {
			line[currentIndex] = scanner.Text()
			currentIndex = (currentIndex + 1) % lines
			itemsAdded++
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error while scanning file: %w", err)
		}
		start := 0
		count := lines
		if itemsAdded < lines {
			count = itemsAdded
		} else {
			start = currentIndex
		}

		for i := range count {
			readIndex := (start + i) % lines
			fmt.Println(line[readIndex])
		}
	}
	return nil
}

func promptFile(path string) bool {
	var answer string
	fmt.Printf("rm: remove '%s'? ", path)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer = strings.ToLower(scanner.Text())
	}

	if answer == "y" || answer == "yes" {
		return true
	}

	fmt.Printf("arquivo '%s' não removido\n", path)
	return false
}

func isReadOnly(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	perm := info.Mode().Perm()

	isProtected := perm&0200 == 0

	return isProtected, nil
}

func mkdirParents(perm int, files string) error {

	dir := strings.Split(files, "/")
	for _, file := range dir {
		err := os.Mkdir(file, os.FileMode(perm))
		if err != nil && os.IsExist(err) {
			return fmt.Errorf("Directory '%s' already exists.", file)
		}
		err = os.Chdir(file)
		if err != nil {
			return fmt.Errorf("Error changing directory: %w", err)
		}

	}
	return nil
}

func rmRecursive(dir string, interactive bool, force bool) error {
	itens, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read dir %s: %w", dir, err)
	}

	for _, item := range itens {
		fullPath := filepath.Join(dir, item.Name())

		if item.IsDir() {
			err := rmRecursive(fullPath, interactive, force)
			if err != nil {
				return fmt.Errorf("aviso: erro no subdiretório %s: %v\n", fullPath, err)
			}
			os.Remove(fullPath)
		} else {
			readOnly, _ := isReadOnly(fullPath)
			if (interactive || readOnly) && !force {
				if promp := promptFile(fullPath); !promp {
					continue
				}
			}
			os.Remove(fullPath)
		}
	}

	return nil

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
		case mode&0111 != 0:
			(*options)[idx] = file + "*"

		}
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	return false
}

func mvPrompt(file string) bool {
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

func lnForce(file string) error {
	if exist := fileExists(file); exist {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("path does not exist: %s", path)
		}
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func Ls(dir string, allDir bool, column bool, classify bool) {
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
		if !strings.HasPrefix(file.Name(), ".") || allDir {
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

	for _, i := range result {
		fmt.Println(i)
	}

}

func Mkdir(perm int, parents bool, dir []string) error {
	for _, files := range dir {
		err := os.Mkdir(files, os.FileMode(perm))

		if err != nil { // Só entramos no switch se houver um erro
			switch {
			case os.IsExist(err):
				return fmt.Errorf("Directory '%s' already exists.", files)

			case parents && os.IsNotExist(err):
				err := mkdirParents(perm, files)
				if err != nil {
					return fmt.Errorf("%w", err)
				}

			default:
				return fmt.Errorf("%w", err)
			}
		}

	}
	return nil
}

func Rm(interactive bool, force bool, recursive bool, dir []string) error {

	for _, files := range dir {
		if recursive {
			err := rmRecursive(files, interactive, force)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
			err = os.Remove(files)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
			continue
		}
		readOnly, err := isReadOnly(files)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		if (interactive || readOnly) && !force {
			if promp := promptFile(files); !promp {
				continue
			}
		}

		err = os.Remove(files)
		if err != nil && os.IsExist(err) && !force {
			return fmt.Errorf("Directory '%s' is not empty.", files)
		} else if err != nil && !os.IsExist(err) && !force {
			return fmt.Errorf("Not posible to remove: %w", err)
		}
	}
	return nil
}

func Cat(byte bool, files ...string) error {

	for _, file := range files {
		if byte {
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

func Head(lines int, files ...string) error {
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("cannot open the file %s: %w", file, err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		line := 0

		if len(files) > 1 {
			fmt.Printf("\n\n==> %s <==\n\n", file)
		}
		for scanner.Scan() && line < lines {
			fmt.Println(scanner.Text())
			line++
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading standard input: %w", err)
		}
	}

	return nil
}

func Tail(file string, bytesString string, linesString string, follow bool) error {

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("cannot open the file %s: %w", file, err)

	}
	defer f.Close()
	switch {
	case follow:
		tailFollow(f, linesString, bytesString)
	case bytesString != "0":
		tailBytePrinter(f, bytesString)
	default:
		tailLinePrinter(f, linesString)
	}

	return nil
}

func Cp(src string, dst string) {

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

func Cmp(file1 string, file2 string, verbose bool, quiet bool) (bool, int, error) {

	f1, err := os.Open(file1)
	if err != nil {

		return false, 2, fmt.Errorf("Error opening file: %w", err)

	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, 2, fmt.Errorf("Error opening file: %w", err)
	}
	defer f2.Close()

	b1 := make([]byte, 1)
	b2 := make([]byte, 1)
	newLine := 1

	for i := 1; ; i++ {

		_, err1 := f1.Read(b1)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			switch {
			case err1 == io.EOF && err2 == io.EOF:
				return true, 0, nil

			case err1 == io.EOF:
				if !quiet {
					fmt.Printf("cmp: EOF on %s after byte %d\n", file1, i)
				}
				return false, 1, nil

			case err2 == io.EOF:
				if !quiet {
					fmt.Printf("cmp: EOF on %s after byte %d\n", file2, i)
				}
				return false, 1, nil

			case err1 != nil:
				return false, 2, fmt.Errorf("error reading %s: %w\n", file1, err1)

			case err2 != nil:
				return false, 2, fmt.Errorf("error reading %s: %w", file2, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			if !verbose {
				if !quiet {
					fmt.Printf("%s %s differ: byte %d line %d\n", file1, file2, i, newLine)
				}
				return false, 1, nil
			}
			fmt.Printf("%d %o %o\n", i, b1, b2)
		}

		if bytes.Equal(b1, []byte{'\n'}) && !verbose {
			newLine++
		}

	}

}

func Mv(files []string, interactive bool, force bool) error {
	if len(files) == 2 {
		info, err := os.Stat(files[1])
		if err != nil {
			if os.IsNotExist(err) {
				err := os.Rename(files[0], files[1])
				if err != nil {
					return fmt.Errorf("Error renaming file: %v", err)
				}
				return nil
			} else {
				return fmt.Errorf("Error: %v\n", err)
			}
		} else {
			if info.IsDir() {
				targetPath := filepath.Join(files[1], files[0])
				err := os.Rename(files[0], targetPath)
				if err != nil {
					return fmt.Errorf("Error moving file: %v", err)
				}
				return nil
			}
		}
		if interactive && !force {
			if asw := mvPrompt(files[1]); !asw {
				return nil
			}
		}
		err = os.Rename(files[0], files[1])
		if err != nil {
			return fmt.Errorf("Error renaming file: %v", err)
		}
		return nil
	}

	for i := range len(files) - 1 {
		info, err := os.Stat(files[len(files)-1])
		if err != nil {
			if os.IsNotExist(err) || !info.IsDir() {
				return fmt.Errorf("Dir does not exist: %v", err)
			} else {
				return fmt.Errorf("Error: %v\n", err)
			}
		}

		if interactive && fileExists(filepath.Join(files[len(files)-1], files[i])) && !force {
			if asw := mvPrompt(files[i]); !asw {
				continue
			}

		}

		err = os.Rename(files[i], filepath.Join(files[len(files)-1], files[i]))
		if err != nil {
			return fmt.Errorf("Error moving file: %v", err)
		}
	}

	return nil
}

func Tee(source io.Reader, files []string, appendFlag bool, ignoreInterrupts bool) error {
	destinations := []io.Writer{os.Stdout}

	if ignoreInterrupts {
		signal.Ignore(os.Interrupt)
	}
	for _, file := range files {
		var f *os.File
		var err error
		if appendFlag {
			f, err = os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		} else {
			f, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		}
		defer f.Close()

		if err != nil {
			return fmt.Errorf("Error opening file: %v", err)
		}

		destinations = append(destinations, f)

	}
	multiWriter := io.MultiWriter(destinations...)

	if _, err := io.Copy(multiWriter, source); err != nil {
		return fmt.Errorf("Error copying data to files: %v", err)
	}
	return nil
}

func Ln(files []string, symbolic bool, force bool, logical bool, physical bool) {
	var err error
	isDir, _ := isDirectory(files[len(files)-1])

	if force {
		if isDir {
			for i := range len(files) - 1 {
				lnForce(filepath.Join(files[len(files)-1], files[i]))
			}
		} else {
			lnForce(files[1])
		}
	}

	for i := range len(files) - 1 {

		if symbolic {
			if isDir {
				err = os.Symlink(files[i], filepath.Join(files[len(files)-1], files[i]))
			} else {
				err = os.Symlink(files[0], files[1])
			}
		} else if logical {
			if isDir {
				finalFile, errEval := filepath.EvalSymlinks(files[i])
				if errEval != nil {
					log.Fatal(errEval)
				}
				err = os.Link(finalFile, filepath.Join(files[len(files)-1], files[i]))
			} else {
				finalFile, errEval := filepath.EvalSymlinks(files[0])
				if errEval != nil {
					log.Fatal(errEval)
				}
				err = os.Link(finalFile, files[1])
			}
		} else if physical {
			if isDir {
				err = os.Link(files[i], filepath.Join(files[len(files)-1], files[i]))
			} else {
				err = os.Link(files[0], files[1])
			}
		} else {
			if isDir {
				err = os.Link(files[i], filepath.Join(files[len(files)-1], files[i]))
			} else {
				err = os.Link(files[0], files[1])
			}
		}

		if err != nil {
			fmt.Printf("Erro link: %v\n", err)
		}

	}

}

func Comm(file1 string, file2 string, noCol1 bool, noCol2 bool, noCol3 bool) error {
	f1, err := os.Open(file1)
	if err != nil {
		return fmt.Errorf("Error opening file %s: %v", file1, err)
	}

	f2, err := os.Open(file2)
	if err != nil {
		return fmt.Errorf("Error opening file %s: %v", file2, err)
	}

	scan1 := bufio.NewScanner(f1)
	scan2 := bufio.NewScanner(f2)

	hasLine1 := scan1.Scan()
	hasLine2 := scan2.Scan()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\tBoth\n", file1, file2)
	for hasLine1 || hasLine2 {
		line1, line2 := scan1.Text(), scan2.Text()
		switch {
		case !hasLine2 || (hasLine1 && line1 < line2):
			if !noCol1 {
				fmt.Fprintf(w, "%s\t\n", line1)
			}
			hasLine1 = scan1.Scan()
		case !hasLine1 || (hasLine2 && line2 < line1):
			if !noCol2 {
				fmt.Fprintf(w, "\t%s\t\n", line2)
			}
			hasLine2 = scan2.Scan()
		case hasLine2 && hasLine1 && line1 == line2:
			if !noCol3 {
				fmt.Fprintf(w, "\t\t%s\t\n", line1)
			}
			hasLine1 = scan1.Scan()
			hasLine2 = scan2.Scan()

		}
	}

	w.Flush()

	if err := scan1.Err(); err != nil {
		return fmt.Errorf("Error scanning %s: %w", file1, err)
	}
	if err := scan2.Err(); err != nil {
		return fmt.Errorf("Error scanning %s: %w", file2, err)
	}

	return nil
}

func chownRecursive(physical bool, logical bool, uid int, gid int) fs.WalkDirFunc {
	switch {
	case physical || (!physical && !logical):
		return func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("Error accessing path %s: %v", path, err)
				return nil
			}

			if err := os.Lchown(path, uid, gid); err != nil {
				return fmt.Errorf("Error: Ownership cannot be changed '%s': %w", path, err)
			}
			return nil
		}

	case logical:
		return func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("Error accessing path %s: %v", path, err)
				return nil
			}

			err = os.Chown(path, uid, gid)
			if err != nil {
				return fmt.Errorf("Error: Ownership cannot be changed '%s': %w", path, err)
			}
			return nil
		}
	}
	return nil
}

func Chown(ug string, files []string, noDereference bool, recursive bool, physical bool, logical bool, hybrid bool) error {

	userId := -1
	groupId := -1

	if num := strings.Index(ug, ":"); num != -1 {
		userStr := ug[:num]
		groupStr := ug[num+1:]

		if userStr != "" {
			userInfo, err := osUser.Lookup(userStr)
			if err != nil {
				return fmt.Errorf("Cannot find user '%s': %w", userStr, err)
			}

			uid, err := strconv.Atoi(userInfo.Uid)
			if err != nil {
				return fmt.Errorf("Error converting the Uid '%s' to int: %w", userInfo.Uid, err)
			}
			userId = uid
		}

		if groupStr != "" {
			groupInfo, err := osUser.LookupGroup(groupStr)
			if err != nil {
				return fmt.Errorf("Cannot find group '%s': %w", groupStr, err)
			}

			gid, err := strconv.Atoi(groupInfo.Gid)
			if err != nil {
				return fmt.Errorf("Error converting Gid '%s' to int: %w", groupInfo.Gid, err)
			}
			groupId = gid
		}

	} else {
		userInfo, err := osUser.Lookup(ug)
		if err != nil {
			return fmt.Errorf("Cannot find user '%s': %w", ug, err)
		}
		uid, err := strconv.Atoi(userInfo.Uid)
		if err != nil {
			return fmt.Errorf("Error converting Uid '%s' to int: %w", userInfo.Uid, err)
		}
		userId = uid
	}

	for _, file := range files {
		if recursive {
			if hybrid {
				fileInfo, err := os.Lstat(file)
				if err != nil {
					return fmt.Errorf("Error: %w", err)
				}

				if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
					file, err = filepath.EvalSymlinks(file)
					if err != nil {
						return fmt.Errorf("Error solving link: %w", err)
					}
				}
				walkFunc := chownRecursive(true, false, userId, groupId)
				filepath.WalkDir(file, walkFunc)

			} else {
				walkFunc := chownRecursive(physical, logical, userId, groupId)
				filepath.WalkDir(file, walkFunc)
			}
		} else if noDereference {
			err := os.Lchown(file, userId, groupId)
			if err != nil {
				return fmt.Errorf("Error changing %s ownership: %w", file, err)
			}
		} else {
			err := os.Chown(file, userId, groupId)
			if err != nil {
				return fmt.Errorf("Error changing %s ownership: %w", file, err)
			}
		}
	}
	return nil
}

func getTouchTLayout(timeString string) (string, error) {
	hasDot := strings.Contains(timeString, ".")
	length := len(timeString)

	switch {
	case length == 8 && !hasDot:
		return "01021504", nil

	case length == 10 && !hasDot:
		return "0601021504", nil

	case length == 11 && hasDot:
		return "01021504.05", nil

	case length == 12 && !hasDot:
		return "200601021504", nil

	case length == 13 && hasDot:
		return "0601021504.05", nil

	case length == 15 && hasDot:
		return "200601021504.05", nil

	default:
		return "", fmt.Errorf("Timestamp '%s' is invalid", timeString)
	}
}

func Touch(files []string, noCreate bool, accessOnly bool, modifyOnly bool, date string, timestamp string, reference bool) error {
	var err error
	aTime := time.Now()
	mTime := time.Now()

	if timestamp != "" {

		format, err := getTouchTLayout(timestamp)
		if err != nil {
			return err
		}

		parsedTime, err := time.ParseInLocation(format, timestamp, time.Local)
		if err != nil {
			return fmt.Errorf("Error on parsing the timestamp '%s' with format '%s': %w", timestamp, format, err)
		}

		if len(timestamp) == 8 || len(timestamp) == 11 {
			parsedTime = parsedTime.AddDate(time.Now().Year(), 0, 0)
		}

		if len(timestamp) == 10 || len(timestamp) == 13 {
			year := parsedTime.Year()
			if year >= 69 && year <= 99 {
				parsedTime = parsedTime.AddDate(1900, 0, 0)
			} else if year >= 0 && year <= 68 {
				parsedTime = parsedTime.AddDate(2000, 0, 0)
			}
		}

		aTime, mTime = parsedTime, parsedTime

	} else if date != "" {
		aTime, err = time.Parse("2006-01-02T15:04:05", date)
		if err != nil {
			return fmt.Errorf("Error with time parse: %w", err)
		}
		mTime, err = time.Parse("2006-01-02T15:04:05", date)
		if err != nil {
			return fmt.Errorf("Error with time parse: %w", err)
		}
	} else if reference {
		refFile := files[len(files)-1]
		files = files[:len(files)-1]
		fStat, err := os.Stat(refFile)
		if err != nil {
			return fmt.Errorf("Cannot get '%s' information: %w", refFile, err)
		}

		mTime = fStat.ModTime()
		if stat, ok := fStat.Sys().(*syscall.Stat_t); ok {
			aTime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
		} else {
			aTime = mTime
		}
	}
	if accessOnly {
		mTime = time.Time{}
	}
	if modifyOnly {
		aTime = time.Time{}
	}

	for _, file := range files {
		fExist := fileExists(file)
		if !fExist && !noCreate {
			f, err := os.Create(file)
			if err != nil {
				return fmt.Errorf("Error creating file '%s': %w", file, err)
			}
			f.Close()

		}
		if err := os.Chtimes(file, aTime, mTime); err != nil {
			return fmt.Errorf("Error chaging '%s' time: %w", file, err)
		}
	}
	return nil
}
