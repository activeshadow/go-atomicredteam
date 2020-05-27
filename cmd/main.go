package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"actshad.dev/go-atomicredteam"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func main() {
	app := &cli.App{
		Name:    "atomicredteam",
		Version: atomicredteam.Version,
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
			&cli.StringFlag{
				Name:    "repo",
				Aliases: []string{"r"},
				Value:   "redcanaryco/master",
				Usage:   "Atomic Red Team repo/branch name",
			},
			&cli.StringSliceFlag{
				Name:  "input",
				Usage: "input key=value pairs",
			},
		},
		Action: func(ctx *cli.Context) error {
			tid := ctx.String("technique")

			if tid == "" {
				return cli.Exit("no technique provided", 1)
			}

			name := ctx.String("name")

			if name == "" {
				return cli.Exit("no test name provided", 1)
			}

			repo := ctx.String("repo")
			inputs := ctx.StringSlice("input")

			test, err := atomicredteam.Execute(tid, name, repo, inputs)
			if err != nil {
				return cli.Exit(err, 1)
			}

			plan, _ := yaml.Marshal(test)

			now := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", ".")

			ioutil.WriteFile(fmt.Sprintf("atomic-test-executor-execution-%s.yaml", now), plan, 0644)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}

}
