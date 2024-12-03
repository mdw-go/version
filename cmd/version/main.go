package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mdw-go/exec"
	"github.com/mdw-go/tui/v2"
	"github.com/mdw-go/version"
)

var Version = "dev"

func main() {
	log.SetFlags(0)
	flags := flag.NewFlagSet(fmt.Sprintf("`%s` @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	flags.Usage = func() {
		_, _ = fmt.Fprintf(flags.Output(), "Usage of %s:\n", flags.Name())
		_, _ = fmt.Fprintln(flags.Output(),
			"When executed in a git repo, shows the user a list of incremented tags to choose from.")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])
	refs, err := exec.Run("git show-ref")
	if err != nil {
		log.Fatalln("Failed to find any git refs (are we in a git repository with at least one commit?):", refs, err)
	}
	describe, err := exec.Run("git describe --tags")
	if err != nil {
		tag(tui.New().Prompt("Enter the initial version tag (remember the 'v' prefix): "))
		return
	}
	describe = strings.TrimSpace(describe)
	raw := describe
	dash := strings.Index(describe, "-")
	if dash >= 0 {
		raw = raw[:dash]
	}
	highest, err := version.Parse(raw)
	if err != nil {
		log.Fatalf("Failed to parse version [%s]: %s", describe, err)
	}
	if dash < 0 {
		log.Println("No changes since last version:", highest)
		return
	}
	var (
		major = highest.Increment("major")
		minor = highest.Increment("minor")
		patch = highest.Increment("patch")
		dev   = highest.Increment("dev")
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
	tag(chosenVersion)
	log.Printf("%s -> %s", highest.String(), chosenVersion)
}

func tag(chosenVersion string) {
	output, err := exec.Run(fmt.Sprintf("git tag -a '%s' -m ''", chosenVersion))
	if err != nil {
		log.Fatalln("Could not update version:", output, err)
	}
	updatedTag, err := exec.Run("git describe")
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
