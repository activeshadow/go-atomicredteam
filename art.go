package atomicredteam

import (
	"embed"
	"regexp"
	"strings"
)

var (
	LOCAL   string
	REPO    string
	BUNDLED bool

	AtomicsFolderRegex = regexp.MustCompile(`PathToAtomicsFolder(\\|\/)`)
	BlockQuoteRegex    = regexp.MustCompile(`<\/?blockquote>`)
)

//go:embed include/*
var include embed.FS

func Logo() []byte {
	logo, err := include.ReadFile("include/logo.txt")
	if err != nil {
		panic(err)
	}

	return logo
}

func Techniques() []string {
	var techniques []string

	entries, err := include.ReadDir("include/atomics")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "T") {
			techniques = append(techniques, entry.Name())
		}
	}

	return techniques
}
