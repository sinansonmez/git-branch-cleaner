package branchcleaner

import (
	"errors"
	"fmt"
	"strings"
)

// Options configures how merged branches are detected and deleted.
type Options struct {
	BaseBranch           string
	Remote               string
	DeleteRemote         bool
	DeleteLocal          bool
	Fetch                bool
	DryRun               bool
	KeepPrefixes         []string
	BaseBranchCandidates []string
}

// Result reports which branches were eligible and deleted.
type Result struct {
	BaseBranch    string
	Remote        string
	MatchedRemote []string
	MatchedLocal  []string
	SkippedRemote []string
	SkippedLocal  []string
	DeletedRemote []string
	DeletedLocal  []string
	DryRun        bool
}

// DefaultOptions returns a safe set of defaults.
func DefaultOptions() Options {
	return Options{
		Remote:               "origin",
		DeleteRemote:         true,
		DeleteLocal:          false,
		Fetch:                true,
		DryRun:               false,
		BaseBranchCandidates: []string{"main", "master"},
	}
}

// Clean deletes branches that are already merged into the base branch.
func Clean(repoPath string, opts Options) (Result, error) {
	var res Result
	if repoPath == "" {
		return res, errors.New("repo path is required")
	}

	if _, err := runGit(repoPath, "rev-parse", "--git-dir"); err != nil {
		return res, fmt.Errorf("not a git repo: %w", err)
	}

	if opts.Remote == "" {
		opts.Remote = "origin"
	}

	if opts.Fetch && opts.DeleteRemote {
		if _, err := runGit(repoPath, "fetch", "--prune", opts.Remote); err != nil {
			return res, fmt.Errorf("git fetch --prune %s: %w", opts.Remote, err)
		}
	}

	base, err := resolveBaseBranch(repoPath, opts.Remote, opts.BaseBranch, opts.BaseBranchCandidates)
	if err != nil {
		return res, err
	}

	res.BaseBranch = base
	res.Remote = opts.Remote
	res.DryRun = opts.DryRun

	if opts.DeleteRemote {
		branches, err := listMergedRemote(repoPath, opts.Remote, base)
		if err != nil {
			return res, err
		}
		branches, skipped := filterBranches(branches, opts.KeepPrefixes)
		res.MatchedRemote = branches
		res.SkippedRemote = skipped

		if !opts.DryRun {
			for _, branch := range branches {
				if err := deleteRemoteBranch(repoPath, opts.Remote, branch); err != nil {
					return res, err
				}
				res.DeletedRemote = append(res.DeletedRemote, branch)
			}
		}
	}

	if opts.DeleteLocal {
		branches, err := listMergedLocal(repoPath, base)
		if err != nil {
			return res, err
		}
		branches, skipped := filterBranches(branches, opts.KeepPrefixes)
		res.MatchedLocal = branches
		res.SkippedLocal = skipped

		if !opts.DryRun {
			for _, branch := range branches {
				if err := deleteLocalBranch(repoPath, branch); err != nil {
					return res, err
				}
				res.DeletedLocal = append(res.DeletedLocal, branch)
			}
		}
	}

	return res, nil
}

func resolveBaseBranch(repoPath, remote, explicit string, candidates []string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	names := filterCandidateNames(candidates)
	if len(names) == 0 {
		names = []string{"main", "master"}
	}

	for _, name := range names {
		if refExists(repoPath, fmt.Sprintf("refs/heads/%s", name)) {
			return name, nil
		}
	}
	if remote != "" {
		for _, name := range names {
			if refExists(repoPath, fmt.Sprintf("refs/remotes/%s/%s", remote, name)) {
				return name, nil
			}
		}
	}

	return names[0], nil
}

func listMergedRemote(repoPath, remote, base string) ([]string, error) {
	baseRef := fmt.Sprintf("%s/%s", remote, base)
	out, err := runGit(repoPath, "branch", "-r", "--merged", baseRef)
	if err != nil {
		return nil, fmt.Errorf("list merged remote branches against %s: %w", baseRef, err)
	}

	return parseRemoteBranches(out, remote, base), nil
}

func listMergedLocal(repoPath, base string) ([]string, error) {
	out, err := runGit(repoPath, "branch", "--merged", base)
	if err != nil {
		return nil, fmt.Errorf("list merged local branches against %s: %w", base, err)
	}

	return parseLocalBranches(out, base), nil
}

func parseRemoteBranches(output, remote, base string) []string {
	var branches []string
	prefix := remote + "/"

	for _, line := range strings.Split(output, "\n") {
		branch := strings.TrimSpace(line)
		if branch == "" {
			continue
		}
		if strings.Contains(branch, "->") {
			continue
		}
		if !strings.HasPrefix(branch, prefix) {
			continue
		}

		name := strings.TrimPrefix(branch, prefix)
		if name == base {
			continue
		}
		branches = append(branches, name)
	}

	return branches
}

func parseLocalBranches(output, base string) []string {
	var branches []string

	for _, line := range strings.Split(output, "\n") {
		branch := strings.TrimSpace(line)
		if branch == "" {
			continue
		}
		if strings.HasPrefix(branch, "* ") {
			continue
		}
		if branch == base {
			continue
		}
		branches = append(branches, branch)
	}

	return branches
}

func filterBranches(branches, keepPrefixes []string) ([]string, []string) {
	if len(keepPrefixes) == 0 {
		return branches, nil
	}

	var filtered []string
	var skipped []string
	for _, branch := range branches {
		if shouldKeepBranch(branch, keepPrefixes) {
			skipped = append(skipped, branch)
			continue
		}
		filtered = append(filtered, branch)
	}

	return filtered, skipped
}

func shouldKeepBranch(branch string, keepPrefixes []string) bool {
	for _, prefix := range keepPrefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		if branch == prefix || strings.HasPrefix(branch, prefix+"/") {
			return true
		}
	}
	return false
}

func filterCandidateNames(candidates []string) []string {
	if len(candidates) == 0 {
		return nil
	}

	var filtered []string
	for _, candidate := range candidates {
		name := strings.TrimSpace(candidate)
		if name == "" {
			continue
		}
		filtered = append(filtered, name)
	}

	return filtered
}

func deleteRemoteBranch(repoPath, remote, branch string) error {
	_, err := runGit(repoPath, "push", remote, "--delete", branch)
	if err != nil {
		return fmt.Errorf("delete remote branch %s/%s: %w", remote, branch, err)
	}
	return nil
}

func deleteLocalBranch(repoPath, branch string) error {
	_, err := runGit(repoPath, "branch", "-d", branch)
	if err != nil {
		return fmt.Errorf("delete local branch %s: %w", branch, err)
	}
	return nil
}
