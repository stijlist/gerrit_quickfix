package main

import (
	"os"
	"strings"
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
			"d3cb25fb9f03878ab921b494929bfb6cd6b57fe5": {
				Patchset: 1,
			},
			"42cccad1662b6986efdcd5e37e1ef54321cca977": {
				Patchset: 2,
			},
		},
	}
	if diff := cmp.Diff(want, parseChange(changeBody), cmp.AllowUnexported(revision{})); diff != "" {
		t.Fatalf("parseChange mismatch -want +got: %s", diff)
	}
}

func TestParseComments(t *testing.T) {
	commentsBody, err := os.Open("./comments.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := comments{
		"garnet/go/src/fidlext/fuchsia/hardware/ethernet/ethernet_fake.go": {
			comment{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Id:       "f8a5d0f8_abfa419c",
				Patchset: 1,
				Line:     47,
				Message:  "clientTxFifo, deviceTxFifo will remain unclosed if this fails.",
			},
			comment{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Id:      "19ec82ab_d83cfc82",
				ReplyTo: "f8a5d0f8_abfa419c",

				Patchset: 1,
				Line:     47,
				Message:  "Done",
			},
		},
		"src/connectivity/network/netstack/link/eth/client_test.go": {
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Id:       "cc90824f_2c2a988a",
				ReplyTo:  "",
				Patchset: 1,
				Line:     76,
				Message:  "why don't you read it from it during the call to Close?",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Id:      "5c77417d_45c68db2",
				ReplyTo: "1acb9f5b_16bb4441",

				Patchset: 1,
				Line:     76,
				Message:  "Actually, I take it back, your original comment made sense; if the channel is unbuffered, asserting on the value sent during close communicates clearly and makes a stronger assertion.",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Id:      "1acb9f5b_16bb4441",
				ReplyTo: "b646d243_4b93b507",

				Patchset: 1,
				Line:     76,
				Message:  "If there's no point, then let's phrase the comment in a way that indicates that. Instead it's worded as if making it unbuffered is desirable.",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Id:       "b646d243_4b93b507",
				ReplyTo:  "cc90824f_2c2a988a",
				Patchset: 1,
				Line:     76,
				Message:  "What would be the point?\nYou can't make it unbuffered without doing a read on a separate goroutine, because Close runs on the test goroutine.\nIf you leave the channel buffered, you'd have to make sure the write happens before the read on this goroutine, which is fine, but again, what would be the point?",
			},
		},
	}
	if diff := cmp.Diff(want, parseComments(commentsBody)); diff != "" {
		t.Fatalf("parseComments mismatch -want +got: %s", diff)
	}
}

func TestPrintComments(t *testing.T) {
	t.Run("standalone", func(t *testing.T) {
		comments := comments{
			"garnet/go/src/fidlext/fuchsia/hardware/ethernet/ethernet_fake.go": {
				comment{
					Author: author{
						Name:  "Tamir Duberstein",
						Email: "tamird@google.com",
					},
					Id:       "f8a5d0f8_abfa419c",
					Patchset: 1,
					Line:     47,
					Message:  "clientTxFifo, deviceTxFifo will remain unclosed if this fails.",
				},
			},
		}
		var b strings.Builder
		printComments(&b, comments)
		want := "garnet/go/src/fidlext/fuchsia/hardware/ethernet/ethernet_fake.go:47\n\ttamird@google.com: clientTxFifo, deviceTxFifo will remain unclosed if this fails."
		if diff := cmp.Diff(want, b.String()); diff != "" {
			t.Fatalf("printComments mismatch (-want +got): %s", diff)
		}
	})
}
