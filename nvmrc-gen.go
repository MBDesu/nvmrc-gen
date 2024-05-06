package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Flags struct {
	isCiMode     bool
	isSilentMode bool
	isMaxNode    bool
	isNonLts     bool
}

var blue = color.New(color.FgBlue).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()
var flags Flags
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

var out io.Writer = os.Stdout

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseFlags() {
	ciModePtr := flag.Bool("c", false, "CI mode. Don't prompt for writing of files.")
	nonLtsPtr := flag.Bool("l", false, "Include non-LTS (oddly versioned) Node releases.")
	maxPtr := flag.Bool("m", false, "Get max Node version rather than min Node version.")
	silentModePtr := flag.Bool("s", false, "Silent mode. Output no logs.")
	flag.Parse()
	flags = Flags{*ciModePtr, *silentModePtr, *maxPtr, *nonLtsPtr}
}

func main() {
	parseFlags()

	if flags.isSilentMode {
		out = io.Discard
	}

	lockfile, err := GetLockfile()
	check(err)
	fileParts := strings.Split(lockfile, "/")
	lockfileName := fileParts[len(fileParts)-1]

	fmt.Fprintln(out, bold(blue("[+]   ")), yellow("found"), lockfileName, yellow("lockfile"))

	var MinMax MinMax
	var minMax string
	if flags.isMaxNode {
		MinMax = Max
		minMax = "maximum"
	} else {
		MinMax = Min
		minMax = "minimum"
	}

	satisfyingNodeVersion := GetSatisfyingNodeVersion(MinMax)
	fmt.Fprintln(out)
	fmt.Fprintln(out, bold(blue("[+]   ")), yellow("found"), satisfyingNodeVersion, yellow(minMax), "Node")
	fmt.Fprintln(out)

	if flags.isCiMode || flags.isSilentMode {
		fmt.Fprintln(out, bold(blue("[+] ")), yellow("writing"), bold(".nvmrc"))
		check(err)
	} else {
		fmt.Fprint(out, bold(green("[?] ")), "Write ", bold(blue(".nvmrc")), " with", " "+satisfyingNodeVersion+"? ", bold(green("([y]/n) ")))
		var yn string
		fmt.Scanln(&yn)
		if !(yn == "N" || yn == "n") {
			fmt.Fprintln(out, bold(blue("[+] ")), yellow("writing"), bold(".nvmrc"))
			err = WriteNvmrc(satisfyingNodeVersion)
			check(err)
		}
	}

	os.Exit(0)
}
