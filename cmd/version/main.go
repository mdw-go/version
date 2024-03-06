package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mdwhatcott/exec"
	"github.com/mdwhatcott/tui/v2"
	"github.com/mdwhatcott/version"
)

var Version = "dev"

func main() {
	log.SetFlags(0)
	flags := flag.NewFlagSet(fmt.Sprintf("%s @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(),
			"When executed in a git repo, shows the user a list of incremented tags to choose from.")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])
	refs, err := exec.Run("git show-ref")
	if err != nil {
		log.Fatalln("Failed to find any git refs (are we in a git repository with at least one commit?):", refs, err)
	}
	rawTags, err := exec.Run("git tag")
	if err != nil {
		log.Fatalln("Failed to run 'git tag':", err)
	}
	var versions []version.Number
	for _, tag := range strings.Split(rawTags, "\n") {
		number, err := version.Parse(tag)
		if err == nil {
			versions = append(versions, number)
		}
	}
	version.Sort(versions)
	var highest version.Number
	if len(versions) == 0 {
		highest = version.Number{Prefix: "v", Dev: -1}
	} else {
		highest = versions[len(versions)-1]
	}
	describe, _ := exec.Run("git describe --tags")
	if highest.String() == describe {
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
	output, err := exec.Run(fmt.Sprintf("git tag -a '%s' -m ''", chosen.String()))
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
	if updated.String() != chosen.String() {
		log.Fatalf("Updated version incorrect. Got: [%s] Want: [%s]", updatedTag, chosen.String())
	}
	log.Printf("%s -> %s", highest.String(), chosen.String())
}
