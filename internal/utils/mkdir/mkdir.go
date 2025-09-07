package mkdir

import (
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"strings"
)

func mkdir(perm int, parents bool, dirs []string) {
	if len(dirs) == 0 {
		log.Printf("mkdir: No directory given for creation")
		return
	}

	for _, dirName := range dirs {
		if strings.TrimSpace(dirName) == "" {
			log.Printf("mkdir: Empty directory name or only spaces; ignored.")
			continue
		}

		var err error
		if parents {
			err = os.MkdirAll(dirName, os.FileMode(perm))
		} else {
			err = os.Mkdir(dirName, os.FileMode(perm))
		}
		if err != nil {
			if os.IsExist(err) {
				log.Printf("mkdir: Directory '%s' already exists.", dirName)

			} else {
				log.Printf("mkdir: Error creating '%s': %v", dirName, err)
			}
		}

	}
}

func Exec() {
	mkdirCmd := flag.NewFlagSet("mkdir", flag.ExitOnError)
	permFlag := mkdirCmd.IntP("mode", "m", 0755, "set file mode")
	parentsFlag := mkdirCmd.BoolP("parents", "p", false, "Create any missing intermediate pathname components.")
	mkdirCmd.Parse(os.Args[2:])
	mkdir(*permFlag, *parentsFlag, mkdirCmd.Args())

}
