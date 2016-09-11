package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// URIScheme is a uri scheme
type URIScheme string

const (
	schemeHTTPS URIScheme = "https"
	schemeSSH             = "ssh"
	schemeGit             = "git"
	schemeEmpty           = ""
)

var (
	// ErrInvalidRepoURI is returned when a repo has a URI we can't parse
	ErrInvalidRepoURI = errors.New("Invalid repo URI")

	patterns = map[URIScheme]*regexp.Regexp{
		schemeHTTPS: regexp.MustCompile(`https://(?P<domain>.+)/(?P<owner>.+)/(?P<repo>.+)(\.git)?`),
		schemeSSH:   regexp.MustCompile(`git@(?P<domain>.+):(?P<owner>.+)/(?P<repo>.+)(\.git)?`),
		schemeGit:   regexp.MustCompile(`git://(?P<domain>.+)/(?P<owner>.+)/(?P<repo>.+)(\.git)?`),
		schemeEmpty: regexp.MustCompile(`((?P<domain>.+)/)?(?P<owner>.+)/(?P<repo>.+)(\.git)?`),
	}
	schemeOrder = []URIScheme{schemeSSH, schemeHTTPS, schemeGit, schemeEmpty}
)

// RepoToPath converts a repository name into a go-like import name, which will double as filesystem path
func RepoToPath(uri string) (string, error) {
	var (
		domain, owner, repo string
	)
	for _, scheme := range schemeOrder {
		pattern := patterns[scheme]
		if pattern.MatchString(uri) {
			match := pattern.FindStringSubmatch(uri)
			result := make(map[string]string)
			for i, name := range pattern.SubexpNames() {
				result[name] = match[i]
			}
			domain = result["domain"]
			if domain == "" {
				domain = "github.com"
			}
			owner = result["owner"]
			repo = result["repo"]
			break
		}
	}
	repo = strings.TrimSuffix(repo, ".git")
	if repo == "" {
		return "", ErrInvalidRepoURI
	}
	return fmt.Sprintf("%s/%s/%s", domain, owner, repo), nil
}
