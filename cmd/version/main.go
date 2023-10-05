package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mdwhatcott/exec"
	"github.com/mdwhatcott/tui"
	"github.com/mdwhatcott/version"
)

var Version = "dev"

func main() {
	log.SetFlags(0)
	flags := flag.NewFlagSet(fmt.Sprintf("%s @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	_ = flags.Parse(os.Args[1:])
	previousTag, _ := exec.Run("git describe --tags")
	previous, _ := version.ParseGitDescribe(previousTag)
	if !previous.Dirty {
		log.Println("No changes since last version:", previous)
		return
	}
	patch := previous.Increment("patch")
	minor := previous.Increment("minor")
	major := previous.Increment("major")
	choices := map[string]version.Number{
		patch.String(): patch,
		minor.String(): minor,
		major.String(): major,
	}
	choice := tui.New(os.Stdin, os.Stdout).Select(fmt.Sprintf("tag to succeed %s", previous.String()),
		patch.String(),
		minor.String(),
		major.String(),
	)
	chosen, ok := choices[choice]
	if !ok {
		log.Println("No action taken at this time.")
		return
	}
	_, err := exec.Run(fmt.Sprintf("git tag -a '%s' -m ''", chosen.String()))
	if err != nil {
		log.Fatalln("Could not update version:", err)
	}
	currentTag, _ := exec.Run("git describe")
	current, err := version.ParseGitDescribe(currentTag)
	if err != nil {
		log.Fatal("Could not parse updated version tag:", err)
	}
	if current.String() != chosen.String() {
		log.Fatalf("Updated version incorrect. Got: [%s] Want: [%s]", currentTag, chosen.String())
	}
	log.Printf("%s -> %s", previous.String(), chosen.String())
}
