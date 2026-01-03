package branchcleaner

import (
	"reflect"
	"testing"
)

func TestParseRemoteBranches(t *testing.T) {
	output := "  origin/HEAD -> origin/main\n  origin/main\n  origin/feature1\n  upstream/feature2\n"
	got := parseRemoteBranches(output, "origin", "main")
	want := []string{"feature1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseLocalBranches(t *testing.T) {
	output := "* main\n  feature1\n  main\n  feature2\n"
	got := parseLocalBranches(output, "main")
	want := []string{"feature1", "feature2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestFilterBranches(t *testing.T) {
	branches := []string{"feature1", "develop", "deploy/api", "main", "bugfix/123"}
	got, skipped := filterBranches(branches, []string{"develop", "main", "deploy"})
	want := []string{"feature1", "bugfix/123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	wantSkipped := []string{"develop", "deploy/api", "main"}
	if !reflect.DeepEqual(skipped, wantSkipped) {
		t.Fatalf("expected skipped %v, got %v", wantSkipped, skipped)
	}
}

func TestFilterCandidateNames(t *testing.T) {
	got := filterCandidateNames([]string{"main", " ", "develop", ""})
	want := []string{"main", "develop"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
