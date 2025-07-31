package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	osUser "os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
	"unicode"
)

type fileInfoStruct struct {
	name, perm, owner, group, targetSym string
	numLinks, inode                     uint64
	uid, gid                            uint32
	size                                int64
	sizeKb                              float64
	ctime, mtime, atime                 time.Time
	symbolic                            bool
}
type DirectoryListing map[string][]fs.DirEntry

// Make ls column view
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
		return fmt.Errorf("Error opening file '%s': %w", file, err)

	}

	b := make([]byte, 1)

	for {
		n, err := files.Read(b)

		if err == io.EOF {
			break
		}

		if err != nil {
			files.Close()
			return fmt.Errorf("Error reading byte: %w", err)
		}

		if n > 0 {
			fmt.Printf("%c", b[0])
		}
	}
	files.Close()
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
func classifyVer(dir string, options *[]fileInfoStruct, classify bool) {
	err := os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for idx, file := range *options {
		fi, err := os.Lstat(file.name)
		if err != nil {
			log.Fatal(err)
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			(*options)[idx].name = file.name + "/"
		case mode&fs.ModeSymlink != 0 && classify:
			(*options)[idx].name = file.name + "@"
		case mode&fs.ModeNamedPipe != 0 && classify:
			(*options)[idx].name = file.name + "|"
		case mode&0111 != 0 && classify:
			(*options)[idx].name = file.name + "*"

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

func lsfileinfo(dir, file string, result *[]fileInfoStruct, dereference, hideControlChars bool) error {
	f := fileInfoStruct{name: file}

	if hideControlChars {
		var builder strings.Builder
		for _, r := range f.name {
			if unicode.IsPrint(r) {
				builder.WriteRune(r)
			} else {
				builder.WriteRune('?')
			}
		}

		f.name = builder.String()
	}

	filePath := filepath.Join(dir, file)

	target, isSym, err := isSymbolic(filePath)
	if err != nil {
		return fmt.Errorf("Error dereferencing '%s': %v", file, err)
	}

	if isSym && dereference {
		f.symbolic = true
		f.targetSym = target
		filePath = target
	}

	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		return fmt.Errorf("Error obtaining '%s' information: %v", filePath, err)
	}

	f.size = fileInfo.Size()
	f.sizeKb = float64(f.size) / 1024
	f.perm = fileInfo.Mode().String()

	sysData := fileInfo.Sys()

	stat, ok := sysData.(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("Error obtaining syscall information '%s'.", filePath)
	}

	f.inode = stat.Ino
	f.numLinks = stat.Nlink

	f.uid = stat.Uid
	f.gid = stat.Gid

	u, err := osUser.LookupId(strconv.Itoa(int(f.uid)))
	if err != nil {
		f.owner = strconv.Itoa(int(f.uid))
	} else {
		f.owner = u.Username
	}

	g, err := osUser.LookupGroupId(strconv.Itoa(int(f.gid)))
	if err != nil {
		f.group = strconv.Itoa(int(f.gid))
	} else {
		f.group = g.Name
	}

	f.mtime = fileInfo.ModTime()
	f.ctime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
	f.atime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)

	*result = append(*result, f)

	return nil
}

func lsFormatLine(i fileInfoStruct, omitOwner, omitGroup, kiloSize, ctime, numericUidGid, inode, accessTime bool) string {
	var ownerStr, groupStr, sizeStr, timeStr, inodeStr, nameStr string

	if i.symbolic {
		nameStr = i.name + " -> " + i.targetSym
	} else {
		nameStr = i.name
	}

	if inode {
		inodeStr = strconv.Itoa(int(i.inode))
	} else {
		inodeStr = ""
	}

	if omitOwner {
		ownerStr = ""

	} else if numericUidGid {
		ownerStr = strconv.Itoa(int(i.uid))
	} else {
		ownerStr = i.owner
	}

	if omitGroup {
		groupStr = ""
	} else if numericUidGid {
		groupStr = strconv.Itoa(int(i.gid))
	} else {
		groupStr = i.group
	}

	if kiloSize {
		sizeStr = fmt.Sprintf("%.1fK", i.sizeKb)
	} else {
		sizeStr = fmt.Sprintf("%d", i.size)
	}

	if ctime {
		timeStr = i.ctime.Format("Jan _2 15:04")
	} else if accessTime {
		timeStr = i.atime.Format("Jan _2 15:04")
	} else {
		timeStr = i.mtime.Format("Jan _2 15:04")
	}

	return fmt.Sprintf("%s %s %d %s %s %s %s %s\n",
		inodeStr,
		i.perm,
		i.numLinks,
		ownerStr,
		groupStr,
		sizeStr,
		timeStr,
		nameStr,
	)
}

func lsPrinter(dir string, almostAllDir, allDir, classify, column, longListing, sortSize, kiloSize, streamFormat, omitOwner, omitGroup, ctime, numericUidGid, inode, dereference, onePerLine, sortmTime, indicatorStyle, hideControlChars, reverse, accessTime, noSort bool, files []os.DirEntry) {
	var result []string

	if allDir || noSort {
		result = append(result, ".", "..")
		almostAllDir = true
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") || (almostAllDir || noSort) {
			result = append(result, file.Name())
		}

	}
	if !noSort {
		sort.Slice(result, func(i, j int) bool {
			if reverse {
				return strings.ToLower(result[i]) > strings.ToLower(result[j])
			} else {
				return strings.ToLower(result[i]) < strings.ToLower(result[j])
			}
		})
	}

	var filesInfo []fileInfoStruct
	for _, file := range result {
		lsfileinfo(dir, file, &filesInfo, dereference, hideControlChars)
	}

	if sortSize && !noSort {
		sort.Slice(filesInfo, func(i, j int) bool {
			if reverse {
				return filesInfo[i].size < filesInfo[j].size
			} else {
				return filesInfo[i].size > filesInfo[j].size
			}
		})

	}

	if sortmTime && !noSort {
		sort.Slice(filesInfo, func(i, j int) bool {
			if accessTime {
				if reverse {
					return filesInfo[i].atime.Format("Jan _2 15:04") < filesInfo[j].atime.Format("Jan _2 15:04")
				} else {
					return filesInfo[i].atime.Format("Jan _2 15:04") > filesInfo[j].atime.Format("Jan _2 15:04")
				}
			} else {
				if reverse {
					return filesInfo[i].mtime.Format("Jan _2 15:04") < filesInfo[j].mtime.Format("Jan _2 15:04")
				} else {
					return filesInfo[i].mtime.Format("Jan _2 15:04") > filesInfo[j].mtime.Format("Jan _2 15:04")
				}
			}
		})
	}

	if classify || indicatorStyle {
		classifyVer(dir, &filesInfo, classify)
	}

	// checks if column view is true and print the result in column view
	if column {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 0, 10, ' ', 0)

		columnise(w, result)

		w.Flush()
		return
	}

	for _, i := range filesInfo {

		switch {
		case longListing:
			fmt.Printf("%s", lsFormatLine(i, omitOwner, omitGroup, kiloSize, ctime, numericUidGid, inode, accessTime))
		case streamFormat && inode:
			fmt.Printf("%d %s, ", i.inode, i.name)
		case streamFormat:
			fmt.Printf("%s, ", i.name)
		case inode:
			fmt.Printf("%d %s  ", i.inode, i.name)
		default:
			fmt.Printf(" %s  ", i.name)
		}
		if onePerLine && !streamFormat {
			fmt.Println()
		}
	}
	fmt.Println()
}

func lsRecursive(rootDir string) (DirectoryListing, error) {
	listings := make(DirectoryListing)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				log.Printf("Error reading dir %q: %v\n", path, err)
				return nil
			}
			listings[path] = files
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return listings, nil
}

func Ls(dirs []string, almostAllDir, column, classify, recursive, allDir, longListing, sortSize, kiloSize, streamFormat, omitOwner, omitGroup, ctime, numericUidGid, inode, dereference, onePerLine, sortmTime, indicatorStyle, hideControlChars, reverseSort, accessTime, noSort bool) {
	if len(dirs) < 1 {
		pwd, _ := os.Getwd()
		dirs = append(dirs, pwd)
	}

	for _, dir := range dirs {
		var listings DirectoryListing
		var err error

		if recursive {
			listings, err = lsRecursive(dir)
		} else {
			var files []fs.DirEntry
			files, err = os.ReadDir(dir)
			if err == nil {
				listings = DirectoryListing{dir: files}
			}
		}

		if err != nil {
			log.Printf("Error reading dir %s: %v", dir, err)
			continue
		}

		for path, files := range listings {
			if len(listings) > 1 {
				fmt.Printf("\n\n%s:\n", path)
			}

			lsPrinter(path, almostAllDir, allDir, classify, column, longListing, sortSize, kiloSize, streamFormat, omitOwner, omitGroup, ctime, numericUidGid, inode, dereference, onePerLine, sortmTime, indicatorStyle, hideControlChars, reverseSort, accessTime, noSort, files)
		}
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

	for _, file := range dir {
		if recursive {
			err := rmRecursive(file, interactive, force)
			if err != nil {
				return fmt.Errorf("Error removing files recursively: %w", err)
			}
			err = os.Remove(file)
			if err != nil {
				return fmt.Errorf("Error removing '%s': %w", file, err)
			}
			continue
		}
		readOnly, err := isReadOnly(file)
		if err != nil {
			return fmt.Errorf("Error determining if '%s' is read-only: %w", file, err)
		}
		if (interactive || readOnly) && !force {
			if promp := promptFile(file); !promp {
				continue
			}
		}

		err = os.Remove(file)
		if err != nil && os.IsExist(err) && !force {
			return fmt.Errorf("Directory '%s' is not empty.", file)
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
	if lines <= 0 {
		return fmt.Errorf("Error: number of lines is 0 or negative")
	}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %w", file, err)
		}

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
			f.Close()
			return fmt.Errorf("reading standard input: %w", err)
		}
		f.Close()

	}

	return nil
}

func Tail(file string, bytesString string, linesString string, follow bool) error {

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': %w", file, err)

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

func isSymbolic(filePath string) (string, bool, error) {
	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, fmt.Errorf("File '%s' does not exist.\n", filePath)
		} else {
			return "", false, fmt.Errorf("Error getting file info for '%s': %v\n", filePath, err)
		}
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(filePath)
		if err != nil {
			return "", false, fmt.Errorf("Error reading symlink target: %v\n", err)
		} else {
			return target, true, nil
		}
	} else {
		return "", false, nil
	}
}

func cpCopyFile(src, dst string, preserveAttributes bool) error {

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		srcFile.Close()
		return err
	}

	_, err = io.Copy(destFile, srcFile)

	srcErr := srcFile.Close()
	destErr := destFile.Close()

	if preserveAttributes {
		preserveFileAttributes(src, dst)
	}

	if err != nil {
		return err
	}
	if srcErr != nil {
		return srcErr
	}
	if destErr != nil {
		return destErr
	}

	return nil
}

func preserveFileAttributes(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if err := os.Chmod(dst, info.Mode().Perm()); err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok {
		atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
		mtime := time.Unix(int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec))
		if err := os.Chtimes(dst, atime, mtime); err != nil {
			return err
		}
		if err := os.Chown(dst, int(stat.Uid), int(stat.Gid)); err != nil && !errors.Is(err, syscall.EPERM) {
			return err
		}
	}
	return nil
}

func Cp(files []string, followSymbolic, recursive, dereference, nodereference, preserveAttributes bool) error {

	isDir, _ := isDirectory(files[len(files)-1])

	for i := range len(files) - 1 {
		if target, isSym, err := isSymbolic(files[i]); followSymbolic && isSym {
			if err != nil {
				return err
			}

			files[i] = target
		}

		if recursive {
			filepath.WalkDir(files[i], func(path string, d fs.DirEntry, err error) error {
				rel, err := filepath.Rel(files[i], path)
				if err != nil {
					return err
				}

				if d.IsDir() {
					err := os.MkdirAll(filepath.Join(files[len(files)-1], rel), 0750)
					if err != nil {
						return err
					}

					err = preserveFileAttributes(path, filepath.Join(files[len(files)-1], rel))
					if err != nil {
						log.Println(err)
						return err
					}

				} else {
					target, isSym, err := isSymbolic(path)
					if err != nil {
						return err
					}

					if nodereference && isSym {
						err := os.Symlink(target, filepath.Join(files[len(files)-1], rel))
						if err != nil {
							log.Println(err)
							return fmt.Errorf("%w", err)
						}
						return nil
					}

					if dereference && isSym {
						path = target
					}

					err = cpCopyFile(path, filepath.Join(files[len(files)-1], rel), preserveAttributes)
					if err != nil {
						return err
					}

				}

				return nil
			})

			return nil
		}

		if isDir {
			err := cpCopyFile(files[i], filepath.Join(files[len(files)-1], filepath.Base(files[i])), preserveAttributes)
			if err != nil {
				return err
			}
		} else {

			target, isSym, err := isSymbolic(files[i])
			if err != nil {
				return err
			}

			if nodereference && isSym {
				err := os.Symlink(target, files[len(files)-1])
				if err != nil {
					log.Println(err)
					return fmt.Errorf("%w", err)
				}
				return nil
			}

			err = cpCopyFile(files[i], files[len(files)-1], preserveAttributes)
			if err != nil {
				return err
			}
		}

	}
	return nil
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

		return false, 2, fmt.Errorf("Error opening file '%s': %w", file1, err)

	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, 2, fmt.Errorf("Error opening file '%s': %w", file2, err)
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
			return fmt.Errorf("Error opening file '%s': %w", file, err)
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
		return fmt.Errorf("Error opening file '%s': %w", file1, err)
	}

	f2, err := os.Open(file2)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': %w", file2, err)
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

func Uniq(input string, output string, duplicated bool, unique bool, counter bool, fields uint, chars uint) error {
	lineCounts := make(map[string]int)
	var outputFileBool bool = false
	var textBytes []byte
	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("Error opening file '%s': %w", input, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lineText string
	for scanner.Scan() {
		if fields > 0 {
			textField := strings.Fields(scanner.Text())
			lineText = strings.Join(textField[fields:], " ")
		} else if chars > 0 {
			lineText = scanner.Text()[chars:]
		} else {
			lineText = scanner.Text()
		}
		lineCounts[lineText]++
	}

	if output != "" {
		outputFileBool = true
	}

	for line, count := range lineCounts {
		if counter {
			if outputFileBool {
				textBytes = fmt.Appendf(textBytes, "%d %s\n", count, line)
			} else {
				fmt.Printf("%d %s\n", count, line)
			}
		} else if count > 1 && duplicated {
			if outputFileBool {
				textBytes = fmt.Appendln(textBytes, line)
			} else {
				fmt.Println(line)
			}
		} else if count == 1 && unique {
			if outputFileBool {
				textBytes = fmt.Appendln(textBytes, line)
			} else {
				fmt.Println(line)
			}
		} else if !counter && !duplicated && !unique {
			if outputFileBool {
				textBytes = fmt.Appendln(textBytes, line)
			} else {
				fmt.Println(line)
			}
		}
	}

	if outputFileBool {
		err := os.WriteFile(output, textBytes, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func cutList(list string) (nums [][2]int, err error) {
	fields := strings.Split(list, ",")
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if strings.HasPrefix(f, "-") {
			end, err := strconv.Atoi(strings.TrimPrefix(f, "-"))
			if err != nil {
				return nil, fmt.Errorf("intervalo inválido: %v", f)
			}
			nums = append(nums, [2]int{0, end})
		} else if strings.HasSuffix(f, "-") {
			start, err := strconv.Atoi(strings.TrimSuffix(f, "-"))
			if err != nil {
				return nil, fmt.Errorf("intervalo inválido: %v", f)
			}
			nums = append(nums, [2]int{start - 1, -1})
		} else if strings.Contains(f, "-") {
			parts := strings.SplitN(f, "-", 2)
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("intervalo inválido: %v", f)
			}
			nums = append(nums, [2]int{start - 1, end})
		} else {
			pos, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("número inválido: %v", f)
			}
			nums = append(nums, [2]int{pos - 1, pos})
		}
	}
	return nums, nil
}

func Cut(files []string, characters, fields, delimiter string, separatedOnly bool) error {
	var listSlice [][2]int
	var err error
	var useChar, useField bool

	if characters != "" {
		listSlice, err = cutList(characters)
		if err != nil {
			return err
		}
		useChar = true
	} else if fields != "" {
		listSlice, err = cutList(fields)
		if err != nil {
			return err
		}
		useField = true
	}

	if delimiter == "" {
		delimiter = "\t"
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if useChar {
				for _, rng := range listSlice {
					start := rng[0]
					end := rng[1]
					if start < 0 {
						start = 0
					}
					if end == -1 || end > len(line) {
						end = len(line)
					}
					if start >= len(line) {
						continue
					}
					if end > len(line) {
						end = len(line)
					}
					if start < end {
						fmt.Print(line[start:end])
					}
				}
				fmt.Println()
			} else if useField {
				if separatedOnly && !strings.Contains(line, delimiter) {
					continue
				}
				fields := strings.Split(line, delimiter)
				var output []string
				for _, rng := range listSlice {
					start := rng[0]
					end := rng[1]
					if start < 0 {
						start = 0
					}
					if end == -1 || end > len(fields) {
						end = len(fields)
					}
					if start >= len(fields) {
						continue
					}
					if end > len(fields) {
						end = len(fields)
					}
					if start < end {
						output = append(output, fields[start:end]...)
					}
				}
				fmt.Println(strings.Join(output, delimiter))
			} else {
				fmt.Println(line)
			}
		}
	}
	return nil
}

func moreSearch(input string, current *int, linesBuffer *[]string, totalLines int, caseInsensitive, tag bool) {
	var contains bool
	found := true
	var term string
	if tag {
		term = input
		found = false
	} else if strings.HasPrefix(input, "/") && len(input) > 1 {
		term = input[1:]
		found = false
	}

	for i := *current; i < totalLines; i++ {
		if caseInsensitive {
			contains = strings.Contains(strings.ToLower((*linesBuffer)[i]), strings.ToLower(term))
		} else {
			contains = strings.Contains((*linesBuffer)[i], term)
		}
		if contains {
			*current = i
			found = true
			break
		}
	}
	if !found {
		fmt.Println("\033[31mPattern not found.\033[0m")
	}

}

func execCommand(commandString string) error {

	commands := []string{}
	if strings.Contains(commandString, ";") {
		commands = strings.Split(commandString, ";")
	} else {
		commands = append(commands, commandString)
	}

	for _, command := range commands {
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if len(output) > 0 {
			fmt.Print(string(output))
		}
		if err != nil {
			return err
		}

	}
	return nil
}

func More(files []string, clearFlag, caseInsensitive, squeeze bool, lines int, commandString, tag string) error {
	rows := 24
	if clearFlag {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Error executing clear command: %w", err)
		}
	}

	if lines > 0 {
		rows = lines
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %w", file, err)
		}
		defer f.Close()

		if commandString != "" {
			err := execCommand(commandString)
			if err != nil {
				return fmt.Errorf("Error executing commands: %w", err)
			}
		}

		scanner := bufio.NewScanner(f)
		linesBuffer := []string{}
		for scanner.Scan() {
			linesBuffer = append(linesBuffer, scanner.Text())
		}
		totalLines := len(linesBuffer)
		count, current := 0, 0

		if tag != "" {
			moreSearch(tag, &current, &linesBuffer, totalLines, caseInsensitive, true)
		}

		for current < totalLines {

			if squeeze && linesBuffer[current] == "" && linesBuffer[current+1] == "" {
				current++
				continue
			}

			fmt.Println(linesBuffer[current])
			current++
			count++

			if count == rows-1 {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("--More--")
				input, _ := reader.ReadString('\n')
				fmt.Printf("\033[1A\033[K")
				count = rows - 2
				input = strings.TrimSpace(input)
				moreSearch(input, &current, &linesBuffer, totalLines, caseInsensitive, false)

			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading standard input: %w", err)
		}
	}
	return nil
}
