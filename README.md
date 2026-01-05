# git-branch-cleaner

Small Go library for deleting merged Git branches using the `git` CLI.

## Usage

```go
package main

import (
	"fmt"

	branchcleaner "git-branch-cleaner"
)

func main() {
	opts := branchcleaner.DefaultOptions()
	opts.BaseBranch = "main"
	opts.DeleteRemote = true
	opts.DeleteLocal = false
	opts.KeepPrefixes = []string{"develop", "main", "master", "deploy"}
	opts.BaseBranchCandidates = []string{"develop", "main", "master"}

	res, err := branchcleaner.Clean("/path/to/repo", opts)
	if err != nil {
		panic(err)
	}

	fmt.Printf("deleted remote: %v\n", res.DeletedRemote)
	fmt.Printf("skipped remote: %v\n", res.SkippedRemote)
}
```

## CLI

```bash
go run ./cmd/git-branch-cleaner -repo /path/to/repo \
  -keep-prefixes develop,main,master,deploy \
  -base-candidates develop,main,master \
  -delete-remote=true -delete-local=false
```

## Install (Homebrew)

```bash
brew tap sinansonmez/git-branch-cleaner
git-branch-cleaner -h
```

Flags:

- `-repo`: path to a local git repo (default `.`)
- `-base`: explicit base branch (overrides candidate detection)
- `-base-candidates`: comma-separated base branch candidates
- `-remote`: remote name (default `origin`)
- `-delete-remote`: delete merged remote branches
- `-delete-local`: delete merged local branches
- `-fetch`: fetch/prune before remote cleanup
- `-dry-run`: report without deleting
- `-keep-prefixes`: comma-separated branch prefixes to keep

## Make targets

```bash
make build
make test
make bundle
make archives
make release TAG=v1.0.0
```

`make release` uses `gh release create` and accepts `RELEASE_FLAGS` (for example `RELEASE_FLAGS=--draft`).

## Notes

- The base branch defaults to candidates in `BaseBranchCandidates` (defaults to `main`, then `master`).
- `BaseBranch` overrides automatic detection.
- A `git fetch --prune` runs before remote cleanup when `Fetch` is true.
- Remote deletion uses `git push <remote> --delete <branch>`.
- `KeepPrefixes` skips branches that match a prefix exactly or as `<prefix>/...`.
- Skipped branches are returned in `Result.SkippedRemote` and `Result.SkippedLocal`.
