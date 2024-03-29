package atomicredteam

import (
	"embed"
	"fmt"
	"regexp"
	"strings"
)

var (
	LOCAL   string
	REPO    string
	BUNDLED bool
	TEMPDIR string

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

	entries, err = include.ReadDir("include/custom")
	if err != nil {
		return techniques
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "T") {
			techniques = append(techniques, entry.Name())
		}
	}

	return techniques
}

func Technique(tid string) ([]byte, string, error) {
	// Check for a custom atomic first, then public.
	if body, err := include.ReadFile("include/custom/" + tid + "/" + tid + ".yaml"); err == nil {
		return body, "include/custom/", nil
	}

	if body, err := include.ReadFile("include/custom/" + tid + "/" + tid + ".yml"); err == nil {
		return body, "include/custom/", nil
	}

	if body, err := include.ReadFile("include/atomics/" + tid + "/" + tid + ".yaml"); err == nil {
		return body, "include/atomics/", nil
	}

	if body, err := include.ReadFile("include/atomics/" + tid + "/" + tid + ".yml"); err == nil {
		return body, "include/atomics/", nil
	}

	return nil, "", fmt.Errorf("Atomic Test is not currently bundled")
}

func Markdown(tid string) ([]byte, error) {
	var (
		body []byte
		err  error
	)

	// Check for a custom atomic first, then public.
	body, err = include.ReadFile("include/custom/" + tid + "/" + tid + ".md")
	if err != nil {
		body, err = include.ReadFile("include/atomics/" + tid + "/" + tid + ".md")
		if err != nil {
			return nil, fmt.Errorf("Atomic Test is not currently bundled")
		}
	}

	return body, nil
}
