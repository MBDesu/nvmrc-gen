package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

var blue = color.New(color.FgBlue).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

var out io.Writer = os.Stdout

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ciModePtr := flag.Bool("c", false, "CI mode. Don't prompt for writing of files.")
	silentModePtr := flag.Bool("s", false, "Silent mode. Output no logs.")

	flag.Parse()

	if *silentModePtr {
		out = io.Discard
	}

	lockfile, err := GetLockfile()
	check(err)
	fileParts := strings.Split(lockfile, "/")
	lockfileName := fileParts[len(fileParts)-1]

	fmt.Fprintln(out, bold(blue("[+]   ")), yellow("found"), lockfileName, yellow("lockfile"))

	minNodeVersion := GetMinNodeVersion()

	fmt.Fprintln(out)
	fmt.Fprintln(out, bold(blue("[+]   ")), yellow("found"), minNodeVersion, yellow("minimum Node"))
	fmt.Fprintln(out)

	if *ciModePtr {
		err = WriteNvmrc(minNodeVersion)
		check(err)
	} else {
		fmt.Fprint(out, bold(green("[?] ")), "Write ", bold(blue(".nvmrc")), " with", " "+minNodeVersion+"? ", bold(green("([y]/n) ")))
		var yn string
		fmt.Scanln(&yn)
		if !(yn == "N" || yn == "n") {
			fmt.Fprintln(out, bold(blue("[+] ")), yellow("writing"), bold(".nvmrc"))
			err = WriteNvmrc(minNodeVersion)
			check(err)
		}
	}

	os.Exit(0)
}
