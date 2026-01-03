package branchcleaner

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runGit(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := stdout.String()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(out)
		}
		if msg != "" {
			return out, fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
		}
		return out, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return out, nil
}

func refExists(repoPath, ref string) bool {
	_, err := runGit(repoPath, "show-ref", "--verify", "--quiet", ref)
	return err == nil
}
