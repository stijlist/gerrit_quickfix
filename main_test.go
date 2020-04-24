package main

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseChangeId(t *testing.T) {
	s := []byte("Change-Id: Iab0c7efa01efa672c96c22474525026d57998043")

	if got, want := parseChangeId(s), "Iab0c7efa01efa672c96c22474525026d57998043"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestParseChange(t *testing.T) {
	changeBody, err := os.Open("./change.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := change{
		Id:              "fuchsia~master~I4782f8c71e5de8571af98c5034a94630c7f685f7",
		CurrentRevision: "42cccad1662b6986efdcd5e37e1ef54321cca977",
		Revisions: map[string]revision{
			"d3cb25fb9f03878ab921b494929bfb6cd6b57fe5": revision{
				Patchset: 1,
			},
			"42cccad1662b6986efdcd5e37e1ef54321cca977": revision{
				Patchset: 2,
			},
		},
	}
	if diff := cmp.Diff(want, parseChange(changeBody), cmp.AllowUnexported(revision{})); diff != "" {
		t.Fatalf("parseChange mismatch -want +got: %s", diff)
	}
}
