package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	branchcleaner "git-branch-cleaner"
)

func main() {
	opts := branchcleaner.DefaultOptions()

	var repoPath string
	var baseBranch string
	var remote string
	var deleteRemote bool
	var deleteLocal bool
	var fetch bool
	var dryRun bool
	var keepPrefixes string
	var baseCandidates string

	flag.StringVar(&repoPath, "repo", ".", "path to git repo")
	flag.StringVar(&baseBranch, "base", "", "base branch to compare against")
	flag.StringVar(&remote, "remote", opts.Remote, "remote name")
	flag.BoolVar(&deleteRemote, "delete-remote", opts.DeleteRemote, "delete merged remote branches")
	flag.BoolVar(&deleteLocal, "delete-local", opts.DeleteLocal, "delete merged local branches")
	flag.BoolVar(&fetch, "fetch", opts.Fetch, "fetch and prune before remote cleanup")
	flag.BoolVar(&dryRun, "dry-run", opts.DryRun, "report branches without deleting")
	flag.StringVar(&keepPrefixes, "keep-prefixes", strings.Join(opts.KeepPrefixes, ","), "comma-separated branch prefixes to keep")
	flag.StringVar(&baseCandidates, "base-candidates", strings.Join(opts.BaseBranchCandidates, ","), "comma-separated base branch candidates")
	flag.Parse()

	opts.BaseBranch = baseBranch
	opts.Remote = remote
	opts.DeleteRemote = deleteRemote
	opts.DeleteLocal = deleteLocal
	opts.Fetch = fetch
	opts.DryRun = dryRun
	opts.KeepPrefixes = splitCSV(keepPrefixes)
	opts.BaseBranchCandidates = splitCSV(baseCandidates)

	if err := runInteractive(repoPath, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runInteractive(repoPath string, opts branchcleaner.Options) error {
	reader := bufio.NewReader(os.Stdin)
	currentRepo := repoPath
	currentKeep := strings.Join(opts.KeepPrefixes, ",")
	currentCandidates := strings.Join(opts.BaseBranchCandidates, ",")

	for {
		fmt.Println("Git Branch Cleaner")
		fmt.Println(strings.Repeat("=", 20))

		currentRepo = promptString(reader, "Repo path", currentRepo)
		opts.BaseBranch = promptString(reader, "Base branch (optional)", opts.BaseBranch)
		opts.Remote = promptString(reader, "Remote", opts.Remote)
		opts.DeleteRemote = promptBool(reader, "Delete remote branches", opts.DeleteRemote)
		opts.DeleteLocal = promptBool(reader, "Delete local branches", opts.DeleteLocal)
		opts.Fetch = promptBool(reader, "Fetch & prune before cleanup", opts.Fetch)
		opts.DryRun = promptBool(reader, "Dry run", opts.DryRun)
		currentKeep = promptString(reader, "Keep prefixes (comma-separated)", currentKeep)
		currentCandidates = promptString(reader, "Base candidates (comma-separated)", currentCandidates)

		opts.KeepPrefixes = splitCSV(currentKeep)
		opts.BaseBranchCandidates = splitCSV(currentCandidates)

		if !promptBool(reader, "Run cleanup now", true) {
			return nil
		}

		res, err := branchcleaner.Clean(currentRepo, opts)
		if err != nil {
			fmt.Printf("\nError: %s\n\n", err)
		} else {
			printResult(res, opts)
		}

		if !promptBool(reader, "Run another cleanup", false) {
			return nil
		}
		fmt.Println()
	}
}

func promptString(reader *bufio.Reader, label, current string) string {
	if current != "" {
		fmt.Printf("%s [%s]: ", label, current)
	} else {
		fmt.Printf("%s: ", label)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return current
	}
	return input
}

func promptBool(reader *bufio.Reader, label string, current bool) bool {
	suffix := "y/N"
	if current {
		suffix = "Y/n"
	}
	for {
		fmt.Printf("%s [%s]: ", label, suffix)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "" {
			return current
		}
		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
		fmt.Println("Please enter y or n.")
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}

	return out
}

func printResult(res branchcleaner.Result, opts branchcleaner.Options) {
	fmt.Printf("\nBase branch: %s\n", res.BaseBranch)
	if opts.DeleteRemote {
		fmt.Printf("Remote: %s\n", res.Remote)
	}
	if opts.DryRun {
		fmt.Println("Dry run: enabled")
	}
	fmt.Println()
	if opts.DeleteRemote {
		printList("Matched remote", res.MatchedRemote)
		printList("Skipped remote", res.SkippedRemote)
		printList("Deleted remote", res.DeletedRemote)
		fmt.Println()
	}
	if opts.DeleteLocal {
		printList("Matched local", res.MatchedLocal)
		printList("Skipped local", res.SkippedLocal)
		printList("Deleted local", res.DeletedLocal)
		fmt.Println()
	}
}

func printList(label string, items []string) {
	if len(items) == 0 {
		fmt.Printf("%s: none\n", label)
		return
	}

	fmt.Printf("%s (%d):\n", label, len(items))
	for _, item := range items {
		fmt.Printf("  - %s\n", item)
	}
}
