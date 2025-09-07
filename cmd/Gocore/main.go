package main

import (
	"fmt"
	"github.com/xelckis/gocore/internal/utils/cal"
	"github.com/xelckis/gocore/internal/utils/cat"
	"github.com/xelckis/gocore/internal/utils/chown"
	"github.com/xelckis/gocore/internal/utils/head"
	"github.com/xelckis/gocore/internal/utils/mkdir"
	"os"
)

func main() {

	progs := map[string]func(){
		"chown": chown.Exec,
		"cal":   cal.Exec,
		"cat":   cat.Exec,
		"head":  head.Exec,
		"mkdir": mkdir.Exec,
	}

	prog, ok := progs[os.Args[1]]

	if !ok {
		fmt.Printf("Command %s not available.\n", os.Args[1])
		return
	}

	prog()
}
