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
				Id:         "f8a5d0f8_abfa419c",
				Patchset:   1,
				Line:       47,
				Message:    "clientTxFifo, deviceTxFifo will remain unclosed if this fails.",
				Unresolved: true,
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
				Id:         "cc90824f_2c2a988a",
				ReplyTo:    "",
				Patchset:   1,
				Line:       76,
				Message:    "why don't you read it from it during the call to Close?",
				Unresolved: true,
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

				Patchset:   1,
				Line:       76,
				Message:    "If there's no point, then let's phrase the comment in a way that indicates that. Instead it's worded as if making it unbuffered is desirable.",
				Unresolved: true,
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

func TestThreadComments(t *testing.T) {
	t.Run("no threads", func(t *testing.T) {
		want := []comment{
			comment{Id: "abc"},
			comment{Id: "def"},
		}
		if diff := cmp.Diff(want, toposort([]comment{
			comment{Id: "abc"}, comment{Id: "def"},
		})); diff != "" {
			t.Fatalf("thread mismatch (-want +got): %s", diff)
		}
	})
	t.Run("find child before parent", func(t *testing.T) {
		want := []comment{
			comment{Id: "abc"},
			comment{Id: "def", ReplyTo: "abc"},
		}
		if diff := cmp.Diff(want, toposort([]comment{
			comment{Id: "def", ReplyTo: "abc"}, comment{Id: "abc"},
		})); diff != "" {
			t.Fatalf("thread mismatch (-want +got): %s", diff)
		}
	})
	t.Run("find parent before child", func(t *testing.T) {
		want := []comment{
			comment{Id: "abc"},
			comment{Id: "def", ReplyTo: "abc"},
		}
		if diff := cmp.Diff(want, toposort([]comment{
			comment{Id: "abc"}, comment{Id: "def", ReplyTo: "abc"},
		})); diff != "" {
			t.Fatalf("thread mismatch (-want +got): %s", diff)
		}
	})
	t.Run("find grandparent, child, parent", func(t *testing.T) {
		want := []comment{
			comment{Id: "abc"},
			comment{Id: "def", ReplyTo: "abc"},
			comment{Id: "ghi", ReplyTo: "def"},
		}
		if diff := cmp.Diff(want, toposort([]comment{
			comment{Id: "abc"},
			comment{Id: "ghi", ReplyTo: "def"},
			comment{Id: "def", ReplyTo: "abc"},
		})); diff != "" {
			t.Fatalf("thread mismatch (-want +got): %s", diff)
		}
	})
}

func TestFilterResolved(t *testing.T) {
	t.Run("error on unsorted input", func(t *testing.T) {
		if got, err := filterResolved([]comment{
			comment{Id: "abc", ReplyTo: "xyz"},
		}); err == nil {
			t.Fatalf("got filterResolved(...) = (%+v, nil), want (nil, non-nil)", got)
		}
		if got, err := filterResolved([]comment{
			comment{Id: "abc"},
			comment{Id: "ghi", ReplyTo: "def"},
			comment{Id: "def", ReplyTo: "abc"},
		}); err == nil {
			t.Fatalf("got filterResolved(...) = (%+v, nil), want (nil, non-nil)", got)
		}
	})
	t.Run("no resolved threads", func(t *testing.T) {
		want := []comment{
			comment{Id: "abc", Unresolved: true},
			comment{Id: "def", ReplyTo: "abc", Unresolved: true},
		}
		got, err := filterResolved([]comment{
			comment{Id: "abc", Unresolved: true},
			comment{Id: "def", ReplyTo: "abc", Unresolved: true},
		})
		if err != nil {
			t.Fatalf("got filterResolved(...) = _, %s, want _, nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("unresolved mismatch (-want +got): %s", diff)
		}
	})
	t.Run("one resolved thread", func(t *testing.T) {
		want := []comment{}
		got, err := filterResolved([]comment{
			comment{Id: "abc", Unresolved: true},
			comment{Id: "def", ReplyTo: "abc"},
		})
		if err != nil {
			t.Fatalf("got filterResolved(...) = _, %s, want _, nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("unresolved mismatch (-want +got): %s", diff)
		}
	})
	t.Run("resolved and unresolved threads", func(t *testing.T) {
		want := []comment{
			comment{Id: "ghi", Unresolved: true},
			comment{Id: "hjk", ReplyTo: "ghi", Unresolved: true},
		}
		got, err := filterResolved([]comment{
			comment{Id: "abc", Unresolved: true},
			comment{Id: "def", ReplyTo: "abc"},
			comment{Id: "ghi", Unresolved: true},
			comment{Id: "hjk", ReplyTo: "ghi", Unresolved: true},
		})
		if err != nil {
			t.Fatalf("got filterResolved(...) = _, %s, want _, nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("unresolved mismatch (-want +got): %s", diff)
		}
	})
	t.Run("resolved roots among resolved and unresolved", func(t *testing.T) {
		want := []comment{
			comment{Id: "ghi", Unresolved: true},
			comment{Id: "hjk", ReplyTo: "ghi", Unresolved: true},
		}
		got, err := filterResolved([]comment{
			comment{Id: "uvw"},
			comment{Id: "abc", Unresolved: true},
			comment{Id: "def", ReplyTo: "abc"},
			comment{Id: "xyz"},
			comment{Id: "ghi", Unresolved: true},
			comment{Id: "hjk", ReplyTo: "ghi", Unresolved: true},
		})
		if err != nil {
			t.Fatalf("got filterResolved(...) = _, %s, want _, nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("unresolved mismatch (-want +got): %s", diff)
		}
	})
	t.Run("unresolved roots among resolved and unresolved", func(t *testing.T) {
		want := []comment{
			comment{Id: "uvw", Unresolved: true},
			comment{Id: "xyz", Unresolved: true},
			comment{Id: "ghi", Unresolved: true},
			comment{Id: "hjk", ReplyTo: "ghi", Unresolved: true},
		}
		got, err := filterResolved([]comment{
			comment{Id: "uvw", Unresolved: true},
			comment{Id: "abc", Unresolved: true},
			comment{Id: "def", ReplyTo: "abc"},
			comment{Id: "xyz", Unresolved: true},
			comment{Id: "ghi", Unresolved: true},
			comment{Id: "hjk", ReplyTo: "ghi", Unresolved: true},
		})
		if err != nil {
			t.Fatalf("got filterResolved(...) = _, %s, want _, nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("unresolved mismatch (-want +got): %s", diff)
		}
	})
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
		want := "garnet/go/src/fidlext/fuchsia/hardware/ethernet/ethernet_fake.go:47: \n\ttamird@google.com: clientTxFifo, deviceTxFifo will remain unclosed if this fails.\n"
		if diff := cmp.Diff(want, b.String()); diff != "" {
			t.Fatalf("printComments mismatch (-want +got): %s", diff)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		comments :=
			comments{
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
			}
		var b strings.Builder
		printComments(&b, comments)
		want := "garnet/go/src/fidlext/fuchsia/hardware/ethernet/ethernet_fake.go:47: \n\ttamird@google.com: clientTxFifo, deviceTxFifo will remain unclosed if this fails.\n\tstijlist@google.com: Done\n"
		if diff := cmp.Diff(want, b.String()); diff != "" {
			t.Fatalf("printComments mismatch (-want +got): %s", diff)
		}
	})
}
