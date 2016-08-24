package vcs

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/iancmcc/jig/manifest"
)

var (
	// Git is the singleton driver
	Git VCS = &gitVCS{}

	absolute = regexp.MustCompile(`(remote: )?([\w\s]+):\s+()(\d+)()(.*)`)
	relative = regexp.MustCompile(`(remote: )?([\w\s]+):\s+(\d+)% \((\d+)/(\d+)\)(.*)`)
)

// GitVCS is a git driver
type gitVCS struct {
}

func parseProgress(repo string, r io.Reader) <-chan Progress {
	out := make(chan Progress)
	scanner := bufio.NewScanner(r)
	scanner.Split(split)
	go func() {
		seen := map[string]struct{}{}
		for scanner.Scan() {
			var (
				begin bool
				match []string
			)
			if match = relative.FindStringSubmatch(scanner.Text()); match == nil {
				match = absolute.FindStringSubmatch(scanner.Text())
			}
			if len(match) == 0 {
				continue
			}
			op := strings.TrimSpace(match[2])
			if strings.HasPrefix(op, "reused") {
				continue
			}
			cur, _ := strconv.Atoi(match[4])
			max, _ := strconv.Atoi(match[5])
			if _, ok := seen[op]; !ok {
				seen[op] = struct{}{}
				begin = true
			}
			prog := Progress{
				repo,
				begin,
				op,
				cur,
				max,
			}
			out <- prog
		}
		close(out)
	}()
	return out
}

func (g *gitVCS) run(repo, cmd string, args ...string) <-chan Progress {
	command := exec.Command("git", append([]string{cmd, "--progress"}, args...)...)
	progout, _ := command.StderrPipe()
	command.Start()
	return parseProgress(repo, progout)
}

// Clone satisfies the VCS interface
func (g *gitVCS) Clone(r manifest.Repo) <-chan Progress {
	return g.run(r.Repo, "clone", r.Repo)
}

// Pull satisfies the VCS interface
func (g *gitVCS) Pull(r manifest.Repo) <-chan Progress {
	return g.run(r.Repo, "pull", r.Repo)

}

// Checkout satisfies the VCS interface
func (g *gitVCS) Checkout(r manifest.Repo) <-chan Progress {
	return g.run("checkout", r.Repo)
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '' {
		return data[0 : len(data)-1]
	}
	return data
}

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, ''); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil

}
