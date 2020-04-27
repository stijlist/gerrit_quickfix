package main

import (
	"encoding/json"
	"flag"
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
	Author     author
	Id         string
	ReplyTo    string `json:"in_reply_to"`
	Patchset   int    `json:"patch_set"`
	Line       int
	Message    string
	Unresolved bool
}

// Keyed by filename.
type comments map[string][]comment

const changePattern = "https://fuchsia-review.googlesource.com/changes/fuchsia~master~%s?o=CURRENT_REVISION&o=CURRENT_COMMIT"
const commentsPattern = "https://fuchsia-review.googlesource.com/changes/fuchsia~master~%s/comments"

func main() {
	var printResolved bool
	flag.BoolVar(&printResolved, "print_resolved", false, "print comment threads that have been resolved")
	flag.Parse()
	out, err := exec.Command("git", "show", "HEAD").Output()
	if err != nil {
		panic("couldn't invoke git show: " + err.Error())
	}
	changeID := parseChangeId(out)
	// TODO: use change to map patchset # -> commit hash
	// TODO: keep only comments from old patchsets that still apply to the same
	// file/line
	// r, err := http.Get(fmt.Sprintf(changePattern, changeID))
	// change := parseChange(r.Body)
	r, err := http.Get(fmt.Sprintf(commentsPattern, changeID))
	if err != nil {
		panic("couldn't get comments endpoint: " + err.Error())
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		panic("unexpected status: " + http.StatusText(r.StatusCode))
	}
	comments := parseComments(r.Body)
	for file := range comments {
		comments[file] = toposort(comments[file])
		if !printResolved {
			filtered, err := filterResolved(comments[file])
			if err != nil {
				panic(err.Error())
			}
			comments[file] = filtered
		}
	}
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
				// The output format here should match the "generic C compiler"
				// errorformat for Vim quickfix. :help errorformat in Vim for more
				// information.
				fmt.Fprintf(w, "%s:%d: \n", file, c.Line)
				currentLine = c.Line
			}
			fmt.Fprintf(w, "\t%s: %s\n", c.Author.Email, c.Message)
		}
	}
}

// filterResolved returns `comments` sans comments with at least one child that are resolved.
// assumes the input has been toposorted.
func filterResolved(comments []comment) ([]comment, error) {
	out := []comment{}
	var prev comment
	var thread []comment
	// root & resolved => skip
	// root & unresolved => commit previous thread
	// child & resolved => throw away thread
	// child & unresolved => continue thread
	for _, c := range comments {
		if c.ReplyTo != "" && c.ReplyTo != prev.Id {
			return nil, fmt.Errorf("unsorted input: got prev.Id = %s, want %s (in %+v)", prev.Id, c.ReplyTo, comments)
		}
		if c.ReplyTo == "" {
			// We're starting a new thread; commit the previous one.
			out = append(out, thread...)
			thread = nil
		}
		thread = append(thread, c)
		if !c.Unresolved {
			// Throw away the previous thread without committing it.
			thread = nil
		}
		prev = c
	}
	// if we're done, commit the last outstanding thread.
	return append(out, thread...), nil
}

// toposort sorts comments into threads of replies, preserving
// the order of roots, where a root is a comment that isn't
// replying to anything.
func toposort(comments []comment) []comment {
	// assumptions: we'll see every id as an Id once and as a ReplyTo at most once.
	roots := []comment{}
	edges := make(map[string]comment)
	for _, c := range comments {
		if c.ReplyTo != "" {
			edges[c.ReplyTo] = c
		} else {
			roots = append(roots, c)
		}
	}
	out := []comment{}
	for _, root := range roots {
		out = append(out, root)
		for next, ok := edges[root.Id]; ok; next, ok = edges[next.Id] {
			out = append(out, next)
		}
	}
	return out
}
