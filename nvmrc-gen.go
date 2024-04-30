package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var blue = color.New(color.FgBlue).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	lockfile, err := GetLockfile()
	check(err)
	fileParts := strings.Split(lockfile, "/")
	lockfileName := fileParts[len(fileParts)-1]

	fmt.Println(bold(blue("[+]   ")), yellow("found"), lockfileName, yellow("lockfile"))

	minNodeVersion := GetMinNodeVersion()

	fmt.Println()
	fmt.Println(bold(blue("[+]   ")), yellow("found"), minNodeVersion, yellow("minimum Node"))
	fmt.Println()

	fmt.Print(bold(green("[?] ")), "Write ", bold(blue(".nvmrc")), " with", " "+minNodeVersion+"? ", bold(green("([y]/n) ")))
	var yn string
	fmt.Scanln(&yn)
	if !(yn == "N" || yn == "n") {
		fmt.Println(bold(blue("[+] ")), yellow("writing"), bold(".nvmrc"))
		err = WriteNvmrc(minNodeVersion)
		check(err)
	}

	os.Exit(0)
}
