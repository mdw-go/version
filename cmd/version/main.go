package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdw-go/exec"
	"github.com/mdw-go/tui/v2"
	"github.com/mdw-go/version/v2"
)

var Version = "dev"

func main() {
	log.SetFlags(0)
	var from string
	var verbose bool
	flags := flag.NewFlagSet(fmt.Sprintf("`%s` @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	flags.StringVar(&from, "from", "", "If supplied, calculate proposed versions from this version value, otherwise run with output of `git describe --tags`.")
	flags.BoolVar(&verbose, "v", false, "verbose mode")
	flags.Usage = func() {
		_, _ = fmt.Fprintf(flags.Output(), "Usage of %s:\n", flags.Name())
		_, _ = fmt.Fprintln(flags.Output(),
			"When executed in a git repo, shows the user a list of incremented tags to choose from. "+
				"The 'dev' tag includes a 'username', either from an environment variable called 'VERSION_USERNAME', "+
				"the first path element of the current git branch, "+
				"or the current OS username (whichever can be resolved first).")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])
	refs, err := execute(verbose, "git show-ref")
	if err != nil {
		log.Fatalln("Failed to find any git refs (are we in a git repository with at least one commit?):", refs, err)
	}
	if from == "" {
		describe, err := execute(verbose, "git describe --tags")
		if err != nil {
			tag(verbose, tui.New().Prompt("Enter the initial version tag (remember the 'v' prefix): "))
			return
		}
		describe = strings.TrimSpace(describe)
		dash := strings.Index(describe, "-")
		if dash >= 0 {
			describe = describe[:dash]
		}
		if dash < 0 {
			log.Println("No changes since last version:", describe)
			return
		}
		from = describe
	}

	highest, err := version.Parse(from)
	if err != nil {
		log.Fatalf("Failed to parse version [%s]: %s", from, err)
	}

	var (
		major = highest.IncrementMajor()
		minor = highest.IncrementMinor()
		patch = highest.IncrementPatch()
		dev   = highest.IncrementDev(fmt.Sprintf("%s-%d", username(), time.Now().Unix()))
	)
	choices := map[string]version.Number{
		major.String(): major,
		minor.String(): minor,
		patch.String(): patch,
		dev.String():   dev,
	}
	choice := tui.New().Select(fmt.Sprintf("tag to succeed %s", highest.String()),
		major.String(),
		minor.String(),
		patch.String(),
		dev.String(),
	)
	chosen, ok := choices[choice]
	if !ok {
		log.Println("No action taken at this time.")
		return
	}
	chosenVersion := chosen.String()
	tag(verbose, chosenVersion)
	log.Printf("%s -> %s", highest.String(), chosenVersion)
}

func username() string {
	username, ok := os.LookupEnv("VERSION_USERNAME")
	if ok {
		return username
	}
	branch, _ := execute(false, "git branch --show-current")
	if root, _, ok := strings.Cut(branch, "/"); ok {
		return root // mikewhat/some-feature -> mikewhat
	}
	osUser, err := user.Current()
	if err != nil {
		log.Fatalln("Failed to resolve current OS user:", err)
	}
	return osUser.Username
}

func tag(verbose bool, chosenVersion string) {
	output, err := execute(verbose, fmt.Sprintf("git tag -a '%s' -m ''", chosenVersion))
	if err != nil {
		log.Fatalln("Could not update version:", output, err)
	}
	updatedTag, err := execute(verbose, "git describe")
	if err != nil {
		log.Fatalln("Failed to run 'git describe':", err)
	}
	updated, err := version.Parse(updatedTag)
	if err != nil {
		log.Fatalln("Could not parse updated version tag:", err)
	}
	if updated.String() != chosenVersion {
		log.Fatalf("Updated version incorrect. Got: [%s] Want: [%s]", updatedTag, chosenVersion)
	}
}

func execute(verbose bool, command string) (string, error) {
	var writer io.Writer = io.Discard
	if verbose {
		writer = os.Stderr
		log.Println(">>>", command)
	}
	return exec.Run(command, exec.Options.Out(writer))
}
