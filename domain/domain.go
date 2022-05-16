package domain

import "regexp"

type Domain struct {
	Name string
	Tld  string
}

var tldRegex = regexp.MustCompile("^.*\\.(.*)$")

func New(name string) *Domain {
	matches := tldRegex.FindStringSubmatch(name)

	return &Domain{
		Name: name,
		Tld:  matches[1],
	}
}
