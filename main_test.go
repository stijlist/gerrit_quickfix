package main

import (
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
	json := strings.NewReader(`
)]}'
{
  "id": "fuchsia~master~I4782f8c71e5de8571af98c5034a94630c7f685f7",
  "project": "fuchsia",
  "branch": "master",
  "hashtags": [],
  "change_id": "I4782f8c71e5de8571af98c5034a94630c7f685f7",
  "subject": "[netstack] Lazily check eth device status",
  "status": "NEW",
  "created": "2020-04-23 01:06:41.000000000",
  "updated": "2020-04-24 00:32:58.000000000",
  "submit_type": "REBASE_ALWAYS",
  "insertions": 154,
  "deletions": 43,
  "total_comment_count": 21,
  "unresolved_comment_count": 1,
  "has_review_started": true,
  "_number": 382873,
  "owner": {
    "_account_id": 18697
  },
  "current_revision": "42cccad1662b6986efdcd5e37e1ef54321cca977",
  "revisions": {
    "d3cb25fb9f03878ab921b494929bfb6cd6b57fe5": {
      "kind": "REWORK",
      "_number": 1,
      "created": "2020-04-23 01:06:41.000000000",
      "uploader": {
        "_account_id": 18697
      },
      "ref": "refs/changes/73/382873/1",
      "fetch": {
        "http": {
          "url": "https://fuchsia.googlesource.com/fuchsia",
          "ref": "refs/changes/73/382873/1"
        }
      }
    },
    "42cccad1662b6986efdcd5e37e1ef54321cca977": {
      "kind": "REWORK",
      "_number": 2,
      "created": "2020-04-24 00:31:33.000000000",
      "uploader": {
        "_account_id": 18697
      },
      "ref": "refs/changes/73/382873/2",
      "fetch": {
        "http": {
          "url": "https://fuchsia.googlesource.com/fuchsia",
          "ref": "refs/changes/73/382873/2"
        }
      }
    }
  },
  "requirements": []
}`)

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
	if diff := cmp.Diff(want, parseChange(json), cmp.AllowUnexported(revision{})); diff != "" {
		t.Fatalf("parseChange mismatch -want +got: %s", diff)
	}
}
