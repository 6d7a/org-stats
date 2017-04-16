package main

import (
	"fmt"
	"os"

	orgstats "github.com/caarlos0/org-stats"
	"github.com/caarlos0/spin"
	"github.com/urfave/cli"
)

var version = "master"

func main() {
	app := cli.NewApp()
	app.Name = "org-stats"
	app.Version = version
	app.Author = "Carlos Alexandro Becker (caarlos0@gmail.com)"
	app.Usage = "Get the contributor stats summary from all repos of any given organization"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			EnvVar: "GITHUB_TOKEN",
			Name:   "token",
			Usage:  "Your GitHub token",
		},
		cli.StringFlag{
			Name:  "org",
			Usage: "GitHub organization to scan.",
		},
	}
	app.Action = func(c *cli.Context) error {
		var token = c.String("token")
		var org = c.String("org")
		if token == "" {
			return cli.NewExitError("missing github api token", 1)
		}
		if org == "" {
			return cli.NewExitError("missing organization name", 1)
		}
		var spin = spin.New("  \033[36m%s Gathering data for '" + org + "'...\033[m")
		spin.Start()
		allStats, err := orgstats.Gather(token, org)
		spin.Stop()
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		printHighlights(allStats)
		return nil
	}
	app.Run(os.Args)
}

func printHighlights(s orgstats.Stats) {
	data := []struct {
		stats  []orgstats.StatPair
		trophy string
		kind   string
	}{
		{
			stats:  orgstats.Sort(s, orgstats.ExtractCommits),
			trophy: "Commit",
			kind:   "commits",
		}, {
			stats:  orgstats.Sort(s, orgstats.ExtractAdditions),
			trophy: "Lines Added",
			kind:   "lines added",
		}, {
			stats:  orgstats.Sort(s, orgstats.ExtractDeletions),
			trophy: "Housekeeper",
			kind:   "lines removed",
		},
	}
	var emojis = []string{"\U0001f3c6", "\U0001f948", "\U0001f949"}
	for _, d := range data {
		fmt.Printf("\033[1m%s champions are:\033[0m\n", d.trophy)
		for i := 0; i < 3; i++ {
			fmt.Printf(
				"%s %s with %d %s!\n",
				emojis[i],
				d.stats[i].Key,
				d.stats[i].Value,
				d.kind,
			)
		}
		fmt.Printf("\n")
	}
}
