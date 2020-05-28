package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sort"
	"time"

	art "actshad.dev/go-atomicredteam"

	"github.com/charmbracelet/glamour"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func main() {
	app := &cli.App{
		Name:    "goart",
		Usage:   "Standalone Atomic Red Team Executor (written in Go)",
		Version: art.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "technique",
				Aliases: []string{"t"},
				Usage:   "technique ID",
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "test name",
			},
			&cli.IntFlag{
				Name:    "index",
				Aliases: []string{"i"},
				Usage:   "test index",
				Value: -1,
			},
			&cli.StringSliceFlag{
				Name:  "input",
				Usage: "input key=value pairs",
			},
			&cli.StringFlag{
				Name:    "local-atomics-path",
				Aliases: []string{"l"},
				Usage:   "directory containing additional/custom atomic test definitions",
			},
			&cli.StringFlag{
				Name:    "dump-technique",
				Aliases: []string{"d"},
				Usage:   "directory to dump the given technique test config to",
			},
		},
		Action: func(ctx *cli.Context) error {
			if art.REPO == "" {
				art.REPO = ctx.String("repo")
			} else {
				art.BUNDLED = true
			}

			if local := ctx.String("local-atomics-path"); local != "" {
				art.LOCAL = local
			}

			fmt.Println(string(art.MustAsset("logo.txt")))

			var (
				tid = ctx.String("technique")
				name = ctx.String("name")
				index = ctx.Int("index")
				inputs = ctx.StringSlice("input")
			)

			if name != "" && index != -1 {
				return cli.Exit("only provide one of 'name' or 'index' flags", 1)
			}

			if tid == "" {
				filter := make(map[string]struct{})

				listTechniques := func() ([]string, error) {
					var (
						techniques   []string
						descriptions []string
					)

					for technique := range filter {
						techniques = append(techniques, technique)
					}

					sort.Strings(techniques)

					for _, tid := range techniques {
						technique, err := art.GetTechnique(tid)
						if err != nil {
							return nil, fmt.Errorf("unable to get technique %s: %w", tid, err)
						}

						descriptions = append(descriptions, fmt.Sprintf("%s - %s", tid, technique.DisplayName))
					}

					return descriptions, nil
				}

				getLocalTechniques := func() error {
					files, err := ioutil.ReadDir(art.LOCAL)
					if err != nil {
						return fmt.Errorf("unable to read contents of provided local atomics path: %w", err)
					}

					for _, f := range files {
						if f.IsDir() && strings.HasPrefix(f.Name(), "T") {
							filter[f.Name()] = struct{}{}
						}
					}

					return nil
				}

				if art.BUNDLED {
					// Get bundled techniques first.
					for _, asset := range art.AssetNames() {
						tokens := strings.Split(asset, "/")

						if tokens[0] == "atomics" {
							if strings.HasPrefix(tokens[1], "T") {
								filter[tokens[1]] = struct{}{}
							}
						}
					}

					// We want to get local techniques after getting bundled techniques so
					// the local techniques will replace any bundled techniques with the
					// same ID.
					if art.LOCAL != "" {
						if err := getLocalTechniques(); err != nil {
							return cli.Exit(err.Error(), 1)
						}
					}

					descriptions, err := listTechniques()
					if err != nil {
						cli.Exit(err.Error(), 1)
					}

					fmt.Println("Locally Available Techniques:\n")

					for _, desc := range descriptions {
						fmt.Println(desc)
					}

					return nil
				}

				// Even if we're not running in bundled mode, still see if the user
				// wants to load any local techniques.
				if art.LOCAL != "" {
					if err := getLocalTechniques(); err != nil {
						return cli.Exit(err.Error(), 1)
					}

					descriptions, err := listTechniques()
					if err != nil {
						cli.Exit(err.Error(), 1)
					}

					fmt.Println("Locally Available Techniques:\n")

					for _, desc := range descriptions {
						fmt.Println(desc)
					}
				}

				orgBranch := strings.Split(art.REPO, "/")

				if len(orgBranch) != 2 {
					return cli.Exit("repo must be in format <org>/<branch>", 1)
				}

				url := fmt.Sprintf("https://github.com/%s/atomic-red-team/tree/%s/atomics", orgBranch[0], orgBranch[1])

				fmt.Printf("Please see %s for a list of available default techniques", url)

				return nil
			}

			if name == "" && index == -1 {
				if dump := ctx.String("dump-technique"); dump != "" {
					dir, err := art.DumpTechnique(dump, tid)
					if err != nil {
						return cli.Exit("error dumping technique: " + err.Error(), 1)
					}

					fmt.Printf("technique %s files dumped to %s", tid, dir)

					return nil
				}

				technique, err := art.GetTechnique(tid)
				if err != nil {
					return cli.Exit("error getting details for " + tid, 1)
				}

				fmt.Printf("Technique: %s - %s\n", technique.AttackTechnique, technique.DisplayName)
				fmt.Println("Tests:")

				for i, t := range technique.AtomicTests {
					fmt.Printf("  %d. %s\n", i, t.Name)
				}

				if runtime.GOOS != "windows" {
					in, err := art.GetMarkdown(tid)
					if err != nil {
						return cli.Exit("error getting Markdown for " + tid, 1)
					}

					renderer, err := glamour.NewTermRenderer(glamour.WithStylePath("dark"), glamour.WithWordWrap(100))
					if err != nil {
						return cli.Exit("error creating new Markdown renderer", 1)
					}

					out, err := renderer.RenderBytes(in)
					if err != nil {
						return cli.Exit("error rendering Markdown for " + tid, 1)
					}

					fmt.Print(string(out))
				}

				return nil
			}

			test, err := art.Execute(tid, name, index, inputs)
			if err != nil {
				return cli.Exit(err, 1)
			}

			plan, _ := yaml.Marshal(test)

			now := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", ".")

			ioutil.WriteFile(fmt.Sprintf("atomic-test-executor-execution-%s-%s.yaml", tid, now), plan, 0644)

			return nil
		},
	}

	if art.REPO == "" {
		app.Flags = append(
			app.Flags,
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Value:   "redcanaryco/master",
				Usage:   "Atomic Red Team repo/branch name",
			},
		)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}

	fmt.Println()
}
