package chown

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"io/fs"
	"log"
	"os"
	osUser "os/user"
	"path/filepath"
	"strconv"
	"strings"
)

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

func userLookUp(userStr string) (int, error) {
	userInfo, err := osUser.Lookup(userStr)
	if err != nil {
		return -1, fmt.Errorf("Cannot find user '%s': %w", userStr, err)
	}

	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		return -1, fmt.Errorf("Error converting the Uid '%s' to int: %w", userInfo.Uid, err)
	}
	return uid, nil

}

func groupLookUp(groupStr string) (int, error) {
	groupInfo, err := osUser.LookupGroup(groupStr)
	if err != nil {
		return -1, fmt.Errorf("Cannot find group '%s': %w", groupStr, err)
	}

	gid, err := strconv.Atoi(groupInfo.Gid)
	if err != nil {
		return -1, fmt.Errorf("Error converting Gid '%s' to int: %w", groupInfo.Gid, err)
	}
	return gid, nil

}

func parseUserGroup(ug string) (userId, groupId int, err error) {

	userId, groupId = -1, -1

	if num := strings.Index(ug, ":"); num != -1 {

		if userStr := ug[:num]; userStr != "" {
			uid, err := userLookUp(userStr)
			if err != nil {
				return -1, -1, err
			}
			userId = uid
		}

		if groupStr := ug[num+1:]; groupStr != "" {
			gid, err := groupLookUp(groupStr)
			if err != nil {
				return -1, -1, err
			}
			groupId = gid
		}

	} else {
		uid, err := userLookUp(ug)
		if err != nil {
			return -1, -1, err
		}
		userId = uid

	}

	return userId, groupId, nil
}

func hybridFunc(file string, userId, groupId int) (err error) {
	var fileOrigin string
	fileInfo, err := os.Lstat(file)
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}

	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		fileOrigin, err = filepath.EvalSymlinks(file)
		if err != nil {
			return fmt.Errorf("Error solving link: %w", err)
		}
	}

	if err != nil {
		return err
	}
	walkFunc := chownRecursive(true, false, userId, groupId)
	filepath.WalkDir(fileOrigin, walkFunc)

	return nil
}

func chown(ug string, files []string, noDereference, recursive, physical, logical, hybrid bool) error {

	userId, groupId, err := parseUserGroup(ug)
	if err != nil {
		return err
	}

	for _, file := range files {
		if recursive {
			if hybrid {
				err := hybridFunc(file, userId, groupId)
				if err != nil {
					return err
				}
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

func Exec() {
	chownCmd := flag.NewFlagSet("chown", flag.ExitOnError)
	noDereferenceFlag := chownCmd.BoolP("no-dereference", "d", false, "affect symbolic links instead of any referenced file (useful only on systems that can change the ownership of a symlink)")
	recursiveFlag := chownCmd.BoolP("reccursive", "R", false, "operate on files and directories recursively")
	physicalFlag := chownCmd.BoolP("physical", "P", false, "do not traverse any symbolic links")
	logicalFlag := chownCmd.BoolP("logical", "L", false, "traverse every symbolic link to a directory encountered")
	hybridFlag := chownCmd.BoolP("Hybrid", "H", false, "if a command line argument is a symbolic link to a directory, traverse it")
	chownCmd.Parse(os.Args[2:])
	err := chown(chownCmd.Arg(0), chownCmd.Args()[1:], *noDereferenceFlag, *recursiveFlag, *physicalFlag, *logicalFlag, *hybridFlag)

	if err != nil {
		fmt.Println(err)
	}

}
