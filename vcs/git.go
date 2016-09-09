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

	"github.com/Sirupsen/logrus"
	"github.com/iancmcc/jig/config"
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

func (g *gitVCS) run(repo, wd string, progress bool, cmd string, args ...string) <-chan Progress {
	if progress {
		args = append([]string{cmd, "--progress"}, args...)
	} else {
		args = append([]string{cmd}, args...)

	}
	strcmd := strings.Join(append([]string{"git"}, args...), " ")
	short, e := RepoToPath(repo)
	if e != nil {
		short = repo
	}
	log := logrus.WithFields(logrus.Fields{
		"cmd":  strcmd,
		"path": wd,
		"repo": short,
	})
	log.Debug("Executing git command")
	var out bytes.Buffer
	command := exec.Command("git", args...)
	command.Dir = wd
	command.Stdout = &out
	progout, _ := command.StderrPipe()
	command.Start()
	result, done := parseProgress(repo, progout)
	go func() {
		<-done
		if err := command.Wait(); err != nil {
			log.WithError(err).Error("Problem running git command")
		}
	}()
	return result
}

func (g *gitVCS) runNoProgress(repo, wd string, args ...string) ([]byte, error) {
	strcmd := strings.Join(append([]string{"git"}, args...), " ")
	short, e := RepoToPath(repo)
	if e != nil {
		short = repo
	}
	log := logrus.WithFields(logrus.Fields{
		"cmd":  strcmd,
		"path": wd,
		"repo": short,
	})
	log.Debug("Executing git command")
	command := exec.Command("git", args...)
	command.Dir = wd
	data, err := command.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(data), nil
}

func prepareDir(dir string) error {
	return os.MkdirAll(filepath.Dir(dir), os.ModeDir|0775)
}

// Clone satisfies the VCS interface
func (g *gitVCS) Clone(r *config.Repo, dir string) (<-chan Progress, error) {
	log := logrus.WithFields(logrus.Fields{
		"repo": r.Repo,
		"ref":  r.Ref,
	})
	log.Debug("Cloning git repo")
	defer log.Debug("Cloned git repo")
	if err := prepareDir(dir); err != nil {
		return nil, err
	}
	out := make(chan Progress)
	go func() {
		defer close(out)
		for p := range g.run(r.Repo, ".", true, "clone", r.Repo, dir) {
			out <- p
		}
		for p := range g.run(r.Repo, dir, true, "fetch", "--all") {
			out <- p
		}
		g.run(r.Repo, dir, false, "branch", "--track", "develop", "origin/develop")
		g.run(r.Repo, dir, false, "branch", "--track", "master", "origin/master")
		g.run(r.Repo, dir, false, "flow", "init", "-d")
	}()
	return out, nil
}

func (g *gitVCS) branch(r *config.Repo, dir string) ([]byte, bool, error) {
	branch, err := g.runNoProgress(r.Repo, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, false, err
	}
	if string(branch) != "HEAD" {
		return branch, true, nil
	}
	// It's a tag. Try to find the best tag
	branch, err = g.runNoProgress(r.Repo, dir, "describe", "--exact-match", "--tags")
	if err != nil {
		branch, err = g.runNoProgress(r.Repo, dir, "rev-parse", "--short", "HEAD")
		if err != nil {
			return nil, false, err
		}
	}
	return branch, false, nil
}

func (g *gitVCS) Status(r *config.Repo, dir string) (*Status, error) {
	branch, _, err := g.branch(r, dir)
	if err != nil {
		return nil, err
	}
	status, err := g.runNoProgress(r.Repo, dir, "status", "-z", "--porcelain")
	if err != nil {
		return nil, err
	}
	short, err := RepoToPath(r.Repo)
	if err != nil {
		return nil, err
	}
	result := &Status{
		Branch:  strings.TrimSpace(string(branch)),
		OrigRef: r.Ref,
		Repo:    short,
	}
	for _, s := range bytes.Split(status, []byte{'\x00'}) {
		if len(s) == 0 {
			continue
		}
		t := s[:2]
		if bytes.HasPrefix(t, []byte("?")) {
			result.Untracked = true
			continue
		}
		if bytes.HasPrefix(t, []byte(" ")) {
			result.Unstaged = true
			continue
		}
		if len(s) > 0 {
			result.Staged = true
		}
	}
	return result, nil
}

// Pull satisfies the VCS interface
func (g *gitVCS) Pull(r *config.Repo, dir string) (<-chan Progress, error) {
	_, isbranch, _ := g.branch(r, dir)
	log := logrus.WithFields(logrus.Fields{
		"repo": r.Repo,
	})
	if !isbranch {
		log.Debug("Skipping pull since not on a branch")
		p := make(chan Progress)
		close(p)
		return p, nil
	}
	log.Debug("Pulling git repo")
	defer log.Debug("Pulled git repo")
	return g.run(r.Repo, dir, true, "pull"), nil
}

// Checkout satisfies the VCS interface
func (g *gitVCS) Checkout(r *config.Repo, dir string) (<-chan Progress, error) {
	return g.run(r.Repo, dir, false, "checkout", r.Ref), nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data = data[0 : len(data)-1]
	}
	if len(data) > 0 && data[len(data)-3] == '' {
		data = data[0 : len(data)-3]
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
