package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
)

type revision struct {
	Patchset int `json:"_number"`
}

type change struct {
	Id              string              `json:"id"`
	CurrentRevision string              `json:"current_revision"`
	Revisions       map[string]revision `json:"revisions"`
}

type author struct {
	Name  string
	Email string
}

type comment struct {
	Author   author
	Id       string
	ReplyTo  string `json:"in_reply_to"`
	Patchset int    `json:"patch_set"`
	Line     int
	Message  string
}

// Keyed by filename.
type comments map[string][]comment

const changePattern = "https://fuchsia-review.googlesource.com/changes/fuchsia~master~%s?o=CURRENT_REVISION&o=CURRENT_COMMIT"
const commentsPattern = "https://fuchsia-review.googlesource.com/chnges/fuchsia~master~%s/comments"

func main() {
	out, err := exec.Command("git show HEAD").Output()
	if err != nil {
		panic("couldn't invoke git show: " + err.Error())
	}
	changeID := parseChangeId(out)
	// r := http.Get(fmt.Sprintf(changePattern, changeID))
	// change := parseChange(r.Body)
	r, err := http.Get(fmt.Sprintf(commentsPattern, changeID))
	if err != nil {
		panic("couldn't get comments endpoint: " + err.Error())
	}
	comments := parseComments(r.Body)

	printComments(os.Stdout, comments)
}

func parseChangeId(git []byte) string {
	return string(regexp.MustCompile("Change-Id: ([[:alnum:]]+)").FindSubmatch(git)[1])
}

const prefix = ")]}'"

func skipPrefix(r io.Reader, p string) (io.Reader, error) {
	discard := make([]byte, len(p))
	n, err := r.Read(discard)
	if err != nil {
		return r, fmt.Errorf("skipPrefix: failed to read: %w", err)
	}
	if n != len(p) {
		return r, fmt.Errorf("skipPrefix: sanitize: got (io.Reader).Read = %d, _, want %d, _", n, len(p))
	}
	if string(discard) != p {
		return r, fmt.Errorf("skipPrefix: got unexpected prefix %s, want %s", discard, p)
	}
	return r, nil
}

func parseChange(r io.Reader) change {
	r, err := skipPrefix(r, prefix)
	if err != nil {
		panic(err)
	}
	var change change
	if err := json.NewDecoder(r).Decode(&change); err != nil {
		panic("couldn't decode JSON: " + err.Error())
	}
	return change
}

func parseComments(r io.Reader) comments {
	r, err := skipPrefix(r, prefix)
	if err != nil {
		panic(err)
	}
	var comments comments
	if err := json.NewDecoder(r).Decode(&comments); err != nil {
		panic("couldn't decode JSON: " + err.Error())
	}
	return comments
}

func printComments(w io.Writer, comments comments) {
	for file, comments := range comments {
		currentLine := -1
		for _, c := range comments {
			if c.Line != currentLine {
				fmt.Fprintf(w, "%s:%d\n", file, c.Line)
				currentLine = c.Line
			}
			fmt.Fprintf(w, "\t%s: %s\n", c.Author.Email, c.Message)
		}
	}
}
