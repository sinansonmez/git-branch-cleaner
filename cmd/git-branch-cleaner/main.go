package main

import (
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
	flag.StringVar(&keepPrefixes, "keep-prefixes", "", "comma-separated branch prefixes to keep")
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

	res, err := branchcleaner.Clean(repoPath, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("base: %s\n", res.BaseBranch)
	if opts.DeleteRemote {
		fmt.Printf("remote: %s\n", res.Remote)
	}
	if opts.DeleteRemote {
		printList("matched remote", res.MatchedRemote)
		printList("skipped remote", res.SkippedRemote)
		printList("deleted remote", res.DeletedRemote)
	}
	if opts.DeleteLocal {
		printList("matched local", res.MatchedLocal)
		printList("skipped local", res.SkippedLocal)
		printList("deleted local", res.DeletedLocal)
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
