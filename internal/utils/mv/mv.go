package mv

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"github.com/xelckis/gocore/internal/utils"
	"os"
	"path/filepath"
)

func mv(paths []string, interactive bool, force bool) error {
	if len(paths) < 2 {
		return fmt.Errorf("mv: need at least one source and one destination")
	}
	info, err := os.Stat(paths[len(paths)-1])
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error: %v\n", err)

	}

	if len(paths) == 2 {
		renameErr := mvRename(info, paths, interactive, force)
		if renameErr != nil {
			return renameErr
		}
		return nil
	}

	if len(paths) > 2 && !info.IsDir() {
		return fmt.Errorf("mv: path does not exist: %v", err)
	}

	for i := range len(paths) - 1 {

		if info.IsDir() && !os.IsNotExist(err) {
			if interactive && utils.FileExists(filepath.Join(paths[len(paths)-1], filepath.Base(paths[i]))) && !force {
				if asw := utils.MvPrompt(paths[i]); !asw {
					continue
				}

			}

			targetPath := filepath.Join(paths[len(paths)-1], filepath.Base(paths[i]))
			err = os.Rename(paths[i], targetPath)
			if err != nil {
				return fmt.Errorf("Error moving file: %v", err)
			}

		} else if !os.IsNotExist(err) {
			return fmt.Errorf("Path does not exist: %v", err)
		}

	}

	return nil
}

func mvRename(info os.FileInfo, paths []string, interactive, force bool) error {
	var target string
	if interactive && utils.FileExists(paths[1]) && !force {
		if asw := utils.MvPrompt(paths[1]); !asw {
			return nil
		}
	}

	if info.IsDir() {
		target = filepath.Join(paths[1], filepath.Base(paths[0]))
	} else {
		target = paths[1]
	}

	err := os.Rename(paths[0], target)
	if err != nil {
		return fmt.Errorf("Error renaming file: %v", err)
	}
	return nil

}

func Exec() {
	mvCmd := flag.NewFlagSet("mv", flag.ExitOnError)
	interactiveFlag := mvCmd.BoolP("interactive", "i", false, "prompt before overwrite")
	forceFlag := mvCmd.BoolP("force", "f", false, "do not prompt before overwriting")
	mvCmd.Parse(os.Args[2:])
	err := mv(mvCmd.Args(), *interactiveFlag, *forceFlag)

	if err != nil {
		fmt.Println(err)
	}

}
