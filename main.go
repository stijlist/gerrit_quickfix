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
	name  string
	email string
}

type comment struct {
	author   author
	patchset int
	line     int
	message  string
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
			if c.line != currentLine {
				fmt.Fprintf(w, "%s:%d\n", file, c.line)
				currentLine = c.line
			}
			fmt.Fprintf(w, "\t%s: %s", c.author.email, c.message)
		}
	}
}
