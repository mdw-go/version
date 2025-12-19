package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mdw-go/exec"
	"golang.org/x/mod/semver"
)

func main() {
	log.SetFlags(0)

	cwd, err := os.Getwd()
	fatalIf(err)

	goModPath, moduleName, err := findGoModule(cwd)
	fatalIf(err)

	gitRoot, _ := findGitRoot()

	prefix, err := moduleTagPrefix(gitRoot, goModPath)
	fatalIf(err)

	if goModPath != "" {
		fmt.Println("Go module:     ", moduleName)
	}
	if prefix == "" {
		log.Println("Module scope:   <root>")
	} else {
		log.Println("Module scope:  ", prefix)
	}

	tags, err := listModuleTags(prefix)
	fatalIf(err)

	latest := latestVersion(tags, prefix)
	latestFull := tagName(prefix, latest)
	if latest == "" {
		latest = prompt("Enter the initial version tag (remember the 'v' prefix): ")
	} else {
		log.Println("Latest version:", latestFull)
	}

	var (
		nextPatch = bumpPatch(latest)
		nextMinor = bumpMinor(latest)
		nextMajor = bumpMajor(latest)
		nextDev   = bumpDev(latest)
	)

	fmt.Println("\nSelect next version:")
	fmt.Println("1) Major:", tagName(prefix, nextMajor))
	fmt.Println("2) Minor:", tagName(prefix, nextMinor))
	fmt.Println("3) Patch:", tagName(prefix, nextPatch))
	fmt.Println("4) Dev:  ", tagName(prefix, nextDev))

	choice := prompt("Choice [1-4]: ")

	var next string
	switch choice {
	case "1":
		next = nextMajor
	case "2":
		next = nextMinor
	case "3":
		next = nextPatch
	case "4":
		next = nextDev
	default:
		fatalIf(errors.New("invalid choice"))
	}

	fullTag := tagName(prefix, next)
	fmt.Println("\nCreating tag:", fullTag)

	log.Printf("module: %s\nversion: %s",
		func() string {
			if prefix == "" {
				return "(root)"
			}
			return prefix
		}(),
		next,
	)
	err = createAnnotatedTag(fullTag)
	fatalIf(err)

	log.Println("Tag created successfully.")
}

/* ---------------- Git & filesystem helpers ---------------- */

func findGoModule(start string) (goMod string, moduleName string, err error) {
	dir := start
	for {
		mod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(mod); err == nil {
			content, err := os.ReadFile(mod)
			if err != nil {
				return "", "", fmt.Errorf("error reading go.mod: %w", err)
			}
			lines := strings.Split(string(content), "\n")
			moduleName = strings.TrimPrefix(lines[0], "module ")
			return mod, moduleName, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", errors.New("no go.mod found")
		}
		dir = parent
	}
}

func findGitRoot() (string, error) {
	out, err := exec.Run("git rev-parse --show-toplevel")
	if err != nil {
		return "", errors.New("not a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func moduleTagPrefix(gitRoot, goModPath string) (string, error) {
	if gitRoot == "" {
		return "", nil
	}
	rel, err := filepath.Rel(gitRoot, filepath.Dir(goModPath))
	if err != nil {
		return "", err
	}
	if rel == "." {
		return "", nil
	}
	return filepath.ToSlash(rel), nil
}

func listModuleTags(prefix string) ([]string, error) {
	pattern := "v*"
	if prefix != "" {
		pattern = prefix + "/v*"
	}

	out, err := exec.Run("git tag --list " + pattern)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
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

/* ---------------- Semver logic ---------------- */

func latestVersion(tags []string, prefix string) string {
	var versions []string

	for _, tag := range tags {
		v := tag
		if prefix != "" {
			v = strings.TrimPrefix(tag, prefix+"/")
		}
		if semver.IsValid(v) {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return ""
	}

	sort.Slice(versions, func(i, j int) bool {
		return semver.Compare(versions[i], versions[j]) > 0
	})

	return versions[0]
}

func bumpDev(latest string) string {
	return fmt.Sprintf("%s-%s-%d", latest, username(), time.Now().Unix())
}

func bumpPatch(v string) string {
	major, minor, patch := split(v)
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch+1)
}

func bumpMinor(v string) string {
	major, minor, _ := split(v)
	return fmt.Sprintf("v%d.%d.0", major, minor+1)
}

func bumpMajor(v string) string {
	major, _, _ := split(v)
	return fmt.Sprintf("v%d.0.0", major+1)
}

func split(v string) (major, minor, patch int) {
	_, _ = fmt.Sscanf(strings.TrimPrefix(v, "v"), "%d.%d.%d", &major, &minor, &patch)
	return
}

/* ---------------- Tag creation ---------------- */

func tagName(prefix, version string) string {
	if prefix == "" {
		return version
	}
	return path.Join(prefix, version)
}

func createAnnotatedTag(tag string) error {
	_, err := exec.Run(fmt.Sprintf("git tag -a %s -m ''", tag))
	return err
}

/* ---------------- UI helpers ---------------- */

func execute(verbose bool, command string) (string, error) {
	var writer io.Writer = io.Discard
	if verbose {
		writer = os.Stderr
		log.Println(">>>", command)
	}
	return exec.Run(command, exec.Options.Out(writer))
}

func prompt(msg string) string {
	_, _ = fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func fatalIf(err error) {
	if err != nil {
		log.Fatalln("error:", err)
	}
}
