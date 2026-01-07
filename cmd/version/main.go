package main

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var Version = "dev"

func main() {
	log.SetFlags(0)

	var fromTag string
	flags := flag.NewFlagSet(fmt.Sprintf("`%s` @ %s", filepath.Base(os.Args[0]), Version), flag.ExitOnError)
	flags.StringVar(&fromTag, "from", "",
		"If supplied, calculate proposed versions from this version value, otherwise run with output of `git describe --tags`.")
	flags.Usage = func() {
		_, _ = fmt.Fprintf(flags.Output(), "Usage of %s:\n", flags.Name())
		_, _ = fmt.Fprintln(flags.Output(),
			"When executed in a git repo, shows the user a list of incremented tags to choose from. "+
				"The 'dev' tag includes a 'username', either from an environment variable called 'VERSION_USERNAME', "+
				"the first path element of the current git branch, "+
				"or the current OS username (whichever can be resolved first). "+
				"When executed in a git repo with multiple Go modules, prompts the user for which module to version.")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])

	rootDir, err := findRootDir()
	if err != nil {
		log.Fatalln("Failed to find the root directory:", err)
	}

	modules, err := findGoModules(rootDir)
	if err != nil {
		log.Fatalln("Failed to find Go modules:", err)
	}

	if len(modules) == 0 {
		log.Fatalln("No Go modules found in the project.")
	}

	var currentModule string
	if len(modules) == 1 {
		currentModule = modules[0]
	} else if len(modules) > 1 {
		for {
			fmt.Println("Please select the desired module:")
			for _, index := range slices.Sorted(maps.Keys(modules)) {
				fmt.Printf("%d. %s\n", index, modules[index])
			}
			choice, _ := strconv.Atoi(prompt("Choice:"))
			selection, ok := modules[choice]
			if ok {
				currentModule = selection
				break
			}
			fmt.Println("Invalid choice, please try again.")
		}
	}

	var latestVersion string
	if fromTag == "" {
		latestVersion = getLatestVersion(rootDir, currentModule)
	} else {
		latestVersion = fromTag
	}

	version := selectVersion(latestVersion, currentModule)
	version = path.Join(currentModule, version)
	err = createGitTag(rootDir, version)
	if err != nil {
		log.Fatalln("Failed to create git tag:", err)
	}
	fmt.Printf("Git tag created: %s\n", version)
}

func selectVersion(latestVersion string, currentModule string) (result string) {
	if latestVersion == "" {
		return prompt("Enter the initial version number (reminder to use a 'v' prefix):")
	}
	fmt.Print("The latest version")
	if currentModule != "" {
		fmt.Print(" for " + currentModule)
	}
	fmt.Printf(": %s\n", latestVersion)

	var prefixV bool
	if latestVersion[0] == 'v' {
		prefixV = true
		latestVersion = latestVersion[1:]
	}

	var versionSelections = map[string]string{
		"1": fmtVersion(prefixV, incrementMajorVersion(latestVersion)),
		"2": fmtVersion(prefixV, incrementMinorVersion(latestVersion)),
		"3": fmtVersion(prefixV, incrementPatchVersion(latestVersion)),
		"4": fmtVersion(prefixV, customVersion(latestVersion)),
	}

	fmt.Println("Please select the next version:")
	fmt.Printf("1. %s\n", versionSelections["1"])
	fmt.Printf("2. %s\n", versionSelections["2"])
	fmt.Printf("3. %s\n", versionSelections["3"])
	fmt.Printf("4. %s\n", versionSelections["4"])
	versionType := prompt("Choice:")

	var ok bool
	for {
		if result, ok = versionSelections[versionType]; ok {
			break
		}
		fmt.Println("Invalid selection, try again.")
	}
	return result
}

func fmtVersion(v bool, version string) string {
	if v {
		return "v" + version
	}
	return version
}

func deriveUsername() string {
	var username string
	defer func() { username = strings.TrimSpace(username) }()
	if username = os.Getenv("VERSION_USERNAME"); username != "" {
		return username
	}
	currentBranch, err := execute("", "git", "branch", "--show-current")
	if err == nil && strings.Contains(currentBranch, "/") {
		dir, _ := path.Split(currentBranch)
		return dir
	}
	osUser, err := user.Current()
	if err != nil {
		log.Fatalln("Failed to resolve current OS user:", err)
	}
	return osUser.Username
}

func findRootDir() (string, error) {
	return execute("", "git", "rev-parse", "--show-toplevel")
}

func findGoModules(rootDir string) (modules map[int]string, err error) {
	output, err := execute("", "find", rootDir, "-name", "go.mod")
	if err != nil {
		return nil, err
	}
	modules = make(map[int]string)
	for i, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, ".")
		line = strings.TrimPrefix(line, rootDir)
		line = strings.TrimPrefix(line, "/")
		line = strings.TrimSuffix(line, "/go.mod")
		modules[i+1] = line
	}
	return modules, nil
}

func getLatestVersion(root, module string) string {
	var output string
	if module == "" {
		output, _ = execute("", "git", "describe", "--tags", "--abbrev=0")
	} else {
		output, _ = execute(root, "git", "tag", "--list", module+"/*")
	}
	tags := strings.Split(output, "\n")
	if len(tags) == 1 && tags[0] == "" {
		return ""
	}
	var toSort []string
	for _, tag := range tags {
		tag = strings.TrimPrefix(tag, module)
		tag = strings.TrimPrefix(tag, "/")
		toSort = append(toSort, tag)
	}
	tags = FilterAndSortSemverTags(toSort)
	output = tags[len(tags)-1]
	output = strings.TrimPrefix(output, module)
	output = strings.TrimPrefix(output, "/")
	return output
}

func incrementMajorVersion(version string) string {
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	return fmt.Sprintf("%d.0.0", major+1)
}
func incrementMinorVersion(version string) string {
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	return fmt.Sprintf("%d.%d.0", major, minor+1)
}
func incrementPatchVersion(version string) string {
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	return fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
}
func customVersion(version string) string {
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	username := deriveUsername()
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d.%d.%d-dev-%s-%d", major, minor, patch, username, timestamp)
}

func createGitTag(root, version string) error {
	_, err := execute(root, "bash", "-c", fmt.Sprintf("git tag -a %s -m ''", version))
	return err
}

func prompt(message string) string {
	fmt.Print(message + " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func execute(dir string, args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

type semver struct {
	major int
	minor int
	patch int
	pre   []string
}

type entry struct {
	original string
	version  semver
}

// FilterAndSortSemverTags filters tags that start with a valid semver
// prefix and returns them sorted by semantic version precedence.
func FilterAndSortSemverTags(tags []string) []string {
	var entries []entry

	for _, tag := range tags {
		if v, ok := parseSemverPrefix(tag); ok {
			entries = append(entries, entry{
				original: tag,
				version:  v,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return compareSemver(entries[i].version, entries[j].version) < 0
	})

	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.original
	}
	return out
}

// parseSemverPrefix parses a semver at the start of s.
func parseSemverPrefix(s string) (semver, bool) {
	var v semver

	if strings.HasPrefix(s, "v") {
		s = s[1:]
	}

	// Split build metadata (ignored for ordering)
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}

	main, pre := s, ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		main = s[:i]
		pre = s[i+1:]
	}

	parts := strings.Split(main, ".")
	if len(parts) != 3 {
		return v, false
	}

	var err error
	if v.major, err = parseInt(parts[0]); err != nil {
		return v, false
	}
	if v.minor, err = parseInt(parts[1]); err != nil {
		return v, false
	}
	if v.patch, err = parseInt(parts[2]); err != nil {
		return v, false
	}

	if pre != "" {
		v.pre = strings.Split(pre, ".")
	}

	return v, true
}

func parseInt(s string) (int, error) {
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return 0, strconv.ErrSyntax
		}
	}
	return strconv.Atoi(s)
}

// compareSemver returns -1 if a < b, 0 if equal, 1 if a > b
func compareSemver(a, b semver) int {
	if a.major != b.major {
		return cmp.Compare(a.major, b.major)
	}
	if a.minor != b.minor {
		return cmp.Compare(a.minor, b.minor)
	}
	if a.patch != b.patch {
		return cmp.Compare(a.patch, b.patch)
	}

	// Handle prerelease precedence
	if len(a.pre) == 0 && len(b.pre) == 0 {
		return 0
	}
	if len(a.pre) == 0 {
		return 1 // release > prerelease
	}
	if len(b.pre) == 0 {
		return -1
	}

	for i := 0; i < len(a.pre) && i < len(b.pre); i++ {
		ai, aNum := numeric(a.pre[i])
		bi, bNum := numeric(b.pre[i])

		switch {
		case aNum && bNum:
			if ai != bi {
				return cmp.Compare(ai, bi)
			}
		case aNum:
			return -1
		case bNum:
			return 1
		default:
			if a.pre[i] != b.pre[i] {
				if a.pre[i] < b.pre[i] {
					return -1
				}
				return 1
			}
		}
	}

	return cmp.Compare(len(a.pre), len(b.pre))
}

func numeric(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil
}
