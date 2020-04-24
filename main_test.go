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
				Patchset: 1,
				Line:     47,
				Message:  "clientTxFifo, deviceTxFifo will remain unclosed if this fails.",
			},
			comment{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     47,
				Message:  "Done",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     60,
				Message:  "unlike the fake, this only has callers in one package. can we put it there?",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     60,
				Message:  "Done",
			},
		},

		"src/connectivity/network/netstack/link/eth/client_test.go": {
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     1,
				Message:  "missing license header",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     1,
				Message:  "Done",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     14,
				Message:  "ordering is inconsistent with convention",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     14,
				Message:  "Done",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     23,
				Message:  "i'm not really understanding the purpose of the test, can you add some words?",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     23,
				Message:  "Done",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     37,
				Message:  "why is this \"NB\"? i believe \"NB\" usually means \"this comment explain why the next line looks weird\" but the next line just initializes a uint32.\n\nfurther, is that actually true? how is the access concurrent?",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     37,
				Message:  "But you're calling `Up` before you call attach. Perhaps the intention was to call attach, wait for the CheckStatus to get \"down\", then set the signal, and then wait for CheckStatus to be called again and receive \"up\"?",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     37,
				Message:  "why the next line looks weird: why use a uint32 for something that's logically a bool?\noh, it's because this is accessed concurrently.\n\nThe comment describes the two methods that access it concurrently; Attach spawns a goroutine that calls CheckStatus.",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     39,
				Message:  "unchecked",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     39,
				Message:  "Done",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     76,
				Message:  "why don't you read it from it during the call to Close?",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     76,
				Message:  "Actually, I take it back, your original comment made sense; if the channel is unbuffered, asserting on the value sent during close communicates clearly and makes a stronger assertion.",
			},
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     76,
				Message:  "If there's no point, then let's phrase the comment in a way that indicates that. Instead it's worded as if making it unbuffered is desirable.",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     76,
				Message:  "What would be the point?\nYou can't make it unbuffered without doing a read on a separate goroutine, because Close runs on the test goroutine.\nIf you leave the channel buffered, you'd have to make sure the write happens before the read on this goroutine, which is fine, but again, what would be the point?",
			},
		},

		"src/connectivity/network/netstack/link/eth/endpoint_test.go": {
			{
				Author: author{
					Name:  "Tamir Duberstein",
					Email: "tamird@google.com",
				},
				Patchset: 1,
				Line:     124,
				Message:  "put this all in one deferred function?",
			},
			{
				Author: author{
					Name:  "Bert Muthalaly",
					Email: "stijlist@google.com",
				},
				Patchset: 1,
				Line:     124,
				Message:  "Done",
			},
		},
	}
	if diff := cmp.Diff(want, parseComments(commentsBody)); diff != "" {
		t.Fatalf("parseComments mismatch -want +got: %s", diff)
	}
}
