package vcs

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

func parseProgress(repo string, r io.Reader) (<-chan Progress, <-chan bool) {
	out := make(chan Progress)
	scanner := bufio.NewScanner(r)
	scanner.Split(split)
	done := make(chan bool)
	go func() {
		seen := map[string]struct{}{}
		for scanner.Scan() {
			var (
				begin bool
				end   bool
				match []string
			)
			text := strings.TrimSpace(scanner.Text())
			if match = relative.FindStringSubmatch(text); match == nil {
				match = absolute.FindStringSubmatch(text)
			}
			if len(match) == 0 {
				continue
			}
			if strings.HasSuffix(text, "done.") {
				end = true
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
				end,
				op,
				cur,
				max,
			}
			out <- prog
		}
		close(out)
		close(done)
	}()
	return out, done
}

func (g *gitVCS) run(repo, wd, cmd string, args ...string) <-chan Progress {
	command := exec.Command("git", append([]string{cmd, "--progress"}, args...)...)
	command.Dir = wd
	progout, _ := command.StderrPipe()
	command.Start()
	result, done := parseProgress(repo, progout)
	go func() {
		<-done
		command.Wait()
	}()
	return result
}

func prepareDir(dir string) error {
	return os.MkdirAll(filepath.Dir(dir), os.ModeDir|0775)
}

// Clone satisfies the VCS interface
func (g *gitVCS) Clone(r manifest.Repo) <-chan Progress {
	dir, err := RepoToPath(r.Repo)
	if err != nil {
		panic(err)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	if err := prepareDir(dir); err != nil {
		panic(err)
	}
	return g.run(r.Repo, ".", "clone", r.Repo, dir)
}

// Pull satisfies the VCS interface
func (g *gitVCS) Pull(r manifest.Repo) <-chan Progress {
	dir, err := RepoToPath(r.Repo)
	if err != nil {
		panic(err)
	}
	return g.run(r.Repo, dir, "pull", r.Repo)

}

// Checkout satisfies the VCS interface
func (g *gitVCS) Checkout(r manifest.Repo) <-chan Progress {
	dir, err := RepoToPath(r.Repo)
	if err != nil {
		panic(err)
	}
	return g.run(r.Repo, dir, "checkout", r.Ref)
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
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
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
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
