package main

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
		"If supplied, calculate proposed versions from this version value, otherwise run with the 'highest' (semver) "+
			"version provided by `git tag`.")
	flags.Usage = func() {
		_, _ = fmt.Fprintf(flags.Output(), "Usage of %s:\n", flags.Name())
		_, _ = fmt.Fprintln(flags.Output(), "TODO")
		flags.PrintDefaults()
	}
	_ = flags.Parse(os.Args[1:])

	root, err := findRootDir()
	if err != nil {
		log.Fatalln("Failed to find git root directory:", err)
	}

	modulePrefix, moduleName := findModuleInfo(root)
	latestVersion := cmp.Or(fromTag, findLatestVersion(modulePrefix))
	selectedVersion := selectVersion(latestVersion, moduleName)

	tagName := selectedVersion
	if modulePrefix != "" {
		tagName = modulePrefix + "/" + selectedVersion
	}

	if err := createGitTag(root, tagName); err != nil {
		log.Fatalln("Failed to create git tag:", err)
	}

	fmt.Printf("Created tag: %s\n", tagName)
}

func selectVersion(latestVersion, currentModule string) (result string) {
	if latestVersion == "" {
		return prompt("Enter the initial version number (reminder to use a 'v' prefix):")
	}
	if currentModule != "" {
		fmt.Println("go module:", currentModule)
	}

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
	fmt.Printf("1. %s -> %s\n", latestVersion, versionSelections["1"])
	fmt.Printf("2. %s -> %s\n", latestVersion, versionSelections["2"])
	fmt.Printf("3. %s -> %s\n", latestVersion, versionSelections["3"])
	fmt.Printf("4. %s -> %s\n", latestVersion, versionSelections["4"])
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
		dir, _, _ := strings.Cut(currentBranch, "/")
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

func findModuleInfo(root string) (prefix, moduleName string) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", ""
	}

	goModPath := filepath.Join(cwd, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return "", ""
	}

	moduleName = readModuleName(goModPath)

	relPath, err := filepath.Rel(root, cwd)
	if err != nil || relPath == "." {
		return "", moduleName
	}

	return relPath, moduleName
}

func readModuleName(goModPath string) string {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func findLatestVersion(modulePrefix string) string {
	output, err := execute("", "git", "tag")
	if err != nil || output == "" {
		return ""
	}

	tags := strings.Split(output, "\n")
	var versions []string

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		version := extractVersion(tag, modulePrefix)
		if version != "" && isSemVer(version) {
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		return ""
	}

	return findHighestVersion(versions)
}

func extractVersion(tag, modulePrefix string) string {
	if modulePrefix == "" {
		return tag
	}
	if strings.HasPrefix(tag, modulePrefix+"/") {
		return strings.TrimPrefix(tag, modulePrefix+"/")
	}
	return ""
}

func isSemVer(version string) bool {
	if version == "" {
		return false
	}
	if version[0] == 'v' {
		version = version[1:]
	}

	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if _, err := parseInt(part); err != nil {
			return false
		}
	}
	return true
}

func findHighestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	highest := versions[0]
	for _, v := range versions[1:] {
		if compareVersions(v, highest) > 0 {
			highest = v
		}
	}
	return highest
}

func compareVersions(a, b string) int {
	aClean, bClean := a, b
	if len(a) > 0 && a[0] == 'v' {
		aClean = a[1:]
	}
	if len(b) > 0 && b[0] == 'v' {
		bClean = b[1:]
	}

	aBase, aPrerelease := splitPrerelease(aClean)
	bBase, bPrerelease := splitPrerelease(bClean)

	aParts := strings.Split(aBase, ".")
	bParts := strings.Split(bBase, ".")

	for i := 0; i < 3; i++ {
		aNum, _ := strconv.Atoi(aParts[i])
		bNum, _ := strconv.Atoi(bParts[i])
		if aNum != bNum {
			return aNum - bNum
		}
	}

	if aPrerelease == "" && bPrerelease != "" {
		return 1
	}
	if aPrerelease != "" && bPrerelease == "" {
		return -1
	}

	return strings.Compare(aPrerelease, bPrerelease)
}

func splitPrerelease(version string) (base, prerelease string) {
	idx := strings.Index(version, "-")
	if idx == -1 {
		return version, ""
	}
	return version[:idx], version[idx+1:]
}
