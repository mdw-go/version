package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mdwhatcott/tui"
	"github.com/mdwhatcott/version"
	"github.com/mdwhatcott/version/git"
)

var Version = "dev"

func main() {
	log.SetFlags(0)
	flags := flag.NewFlagSet(fmt.Sprintf("%s @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	_ = flags.Parse(os.Args[1:])
	repository := new(git.Repository)
	previous, _ := repository.CurrentVersion()
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
	choice := tui.New(os.Stdin, os.Stdout).Select("tag",
		patch.String(),
		minor.String(),
		major.String(),
	)
	chosen, ok := choices[choice]
	if !ok {
		log.Println("No action taken at this time.")
		return
	}
	err := repository.UpdateVersion(chosen)
	if err != nil {
		log.Fatalln("Could not update version:", err)
	}
	log.Printf("%v -> %v", previous, chosen)
}
