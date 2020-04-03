package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	com "github.com/hatchify/mod-common"
	common "github.com/hatchify/mod-common"
	gomu "github.com/hatchify/mod-utils"
	flag "github.com/hatchify/parg"
)

var version = "undefined"
var logLevel = "NORMAL"

func readInput() {
	var (
		err  error
		text string
	)

	files := make([]string, 0)
	reader := bufio.NewReader(os.Stdin)

	// Get files from stdin (piped from another program's output)
	for err == nil {
		if text = strings.TrimSpace(text); len(text) > 0 {
			files = append(files, text)
		}

		text, err = reader.ReadString('\n')
	}

	// Print files
	for i := range files {
		fmt.Println(files[i])
	}
}

func showHelp(cmd *flag.Command) {
	if cmd == nil {
		fmt.Println(flag.Help())
	} else {
		fmt.Println(cmd.ShowHelp())
	}
}

func exitWithError(message string) {
	com.Errorln(message)
	os.Exit(1)
}

// Parg will parse your args
func getCommand() (cmd *flag.Command, err error) {
	// Command/Arg/Flag parser
	parg := flag.New()

	// Configure commands
	parg.AddAction("", "Note - Will accept multiple arguments\n  Aggregate libs to crawl the dependency chain.\n  Providing no arguments will act on files in selected directories\n  (Be Careful!)\n\n  Usage: `gomu <optional flags> cmd args <optional flags>`")
	parg.AddAction("help", "Prints available commands and flags.\n  Use `gomu help <command> <flags>` to get more specific info")
	parg.AddAction("version", "Prints current version. Use ./install.sh to get version support")

	parg.AddAction("list", "Prints each file in dependency chain")
	parg.AddAction("pull", "Updates branch for file in dependency chain.\n  Providing a -branch will checkout given branch.\n  Creates branch if provided none exists.")

	parg.AddAction("replace", "Replaces each versioned file in the dependency chain\n  Uses the current checked out local copy")
	parg.AddAction("reset", "Reverts go.mod and go.sum back to last committed version.\n  Usage: `gomu reset mod-common parg`")

	parg.AddAction("sync", "Updates modfiles\n  Conditionally performs extra tasks depending on flags.\n  Usage: `gomu <flags> sync mod-common parg simply <flags>`")

	// Configure flags
	parg.AddGlobalFlag(flag.Flag{ // Directories to search in
		Name:        "-include",
		Identifiers: []string{"-i", "-in", "-include"},
		Type:        flag.STRINGS,
		Help:        "Will aggregate files in 1 or more directories.\n  Usage: `gomu list -i hatchify -i vroomy`",
	})
	parg.AddGlobalFlag(flag.Flag{ // Branch to checkout/create
		Name:        "-branch",
		Identifiers: []string{"-b", "-branch"},
		Help:        "Will checkout or create said branch\n  Updating or creating a pull request\n  Depending on command and other flags.\n  Usage: `gomu pull -b feature/Jira-Ticket`",
	})
	parg.AddGlobalFlag(flag.Flag{ // Minimal output for | chains
		Name:        "-name-only",
		Identifiers: []string{"-name", "-name-only"},
		Type:        flag.BOOL,
		Help:        "Will reduce output to just the filenames changed\n  (ls-styled output for | chaining)\n  Usage: `gomu list -name`",
	})
	parg.AddGlobalFlag(flag.Flag{ // Commits local changes
		Name:        "-commit",
		Identifiers: []string{"-c", "-commit"},
		Type:        flag.BOOL,
		Help:        "Will commit local changes if present\n  Includes all files outside of mod files\n  Usage: `gomu sync -c`",
	})
	parg.AddGlobalFlag(flag.Flag{ // Creates pull request if possible
		Name:        "-pull-request",
		Identifiers: []string{"-pr", "-pull-request"},
		Type:        flag.BOOL,
		Help:        "Will create a pull request if possible\n  Fails if on master, or if no changes\n  Usage: `gomu sync -pr`",
	})
	parg.AddGlobalFlag(flag.Flag{ // Branch to checkout/create
		Name:        "-message",
		Identifiers: []string{"-m", "-msg", "-message"},
		Help:        "Will set a custom commit message\n  Applies to -c and -pr flags.\n  Usage: `gomu sync -c -m \"Update all the things!\"`",
	})
	parg.AddGlobalFlag(flag.Flag{ // Update tag/version for changed libs or subdeps
		Name:        "-tag",
		Identifiers: []string{"-t", "-tag"},
		Type:        flag.BOOL,
		Help:        "Will increment tag if new commits since last tag\n  Requires tag previously set\n  Usage: `gomu sync -t`",
	})

	return flag.Validate()
}

func gomuOptions() (options gomu.Options) {
	// Get command from args
	cmd, err := getCommand()

	// TODO: cmd.Help() && parg.Help()

	if err != nil {
		// Show usage and exit with error
		showHelp(nil)
		com.Errorln("\nError parsing arguments: ", err)
		os.Exit(1)
	}
	if cmd == nil {
		showHelp(cmd)
		com.Errorln("\nError parsing command: ", err)
		os.Exit(1)
	}

	switch cmd.Action {
	case "version":
		// Print version and exit without error
		fmt.Println(version)
		os.Exit(0)
	case "help", "", " ":
		// Print help and exit without error
		showHelp(cmd)
		os.Exit(0)
	}

	// Parse options from cmd
	options.Action = cmd.Action

	// Args
	options.FilterDependencies = make([]string, len(cmd.Arguments))
	for i, argument := range cmd.Arguments {
		options.FilterDependencies[i] = argument.Name
	}

	// Flags
	options.TargetDirectories = cmd.StringsFrom("-include")

	options.Branch = cmd.StringFrom("-branch")
	options.CommitMessage = cmd.StringFrom("-message")

	options.Commit = cmd.BoolFrom("-commit")
	options.PullRequest = cmd.BoolFrom("-pull-request")
	options.Tag = cmd.BoolFrom("-tag")
	nameOnly := cmd.BoolFrom("-name-only")
	if nameOnly {
		options.LogLevel = com.NAMEONLY
	} else {
		options.LogLevel = com.NORMAL
	}

	return
}

func fromArgs() *gomu.MU {
	options := gomuOptions()
	common.SetLogLevel(options.LogLevel)

	if len(options.TargetDirectories) == 0 {
		options.TargetDirectories = []string{"."}
	}

	return gomu.New(options)
}
