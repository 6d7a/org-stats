package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/caarlos0/duration"
	"github.com/caarlos0/org-stats/csv"
	"github.com/caarlos0/org-stats/highlights"
	orgstats "github.com/caarlos0/org-stats/orgstats"
	"github.com/caarlos0/spin"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	token          string
	organization   string
	githubURL      string
	since          string
	csvPath        string
	blacklist      []string
	top            int
	includeReviews bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "org-stats",
		Short: "Get the contributor stats summary from all repos of any given organization",
		Long: `org-stats can be used to get an overall sense of your org's contributors.

It uses the GitHub API to grab the repositories in the given organization.
Then, iterating one by one, it gets statistics of lines added, removed and number of commits of contributors.
After that, if opted in, it does several searches to get the number of pull requests reviewed by each of the previously find contributors.
Finally, it prints a rank by each category.


Important notes:
* The ` + "`" + `--since` + "`" + ` filter does not work "that well" because GitHub summarizes thedata by week, so the data is not as granular as it should be.
* The ` + "`" + `--include-reviews` + "`" + ` only grabs reviews from users that had contributions on the previous step.
* In the ` + "`" + `--blacklist` + "`" + ` option, 'foo' blacklists both the 'foo' user and 'foo' repo, while 'user:foo' blacklists only the user and 'repo:foo' only the repository.
* The ` + "`" + `--since` + "`" + ` option accepts all the regular time.Durations Go accepts, plus a few more: 1y (365d), 1mo (30d), 1w (7d) and 1d (24h).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			spin := spin.New(loadingStyle.Render("  %s Gathering data for '" + organization + "'..."))
			spin.Start()

			sinceD, err := duration.Parse(since)
			if err != nil {
				return fmt.Errorf("invalid --since duration: '%s'", since)
			}

			userBlacklist, repoBlacklist := buildBlacklists(blacklist)

			allStats, err := orgstats.Gather(
				token,
				organization,
				userBlacklist,
				repoBlacklist,
				githubURL,
				time.Now().UTC().Add(-1*time.Duration(sinceD)),
				includeReviews,
			)
			spin.Stop()
			if err != nil {
				return err
			}

			if csvPath != "" {
				if err := os.MkdirAll(filepath.Dir(csvPath), 0755); err != nil {
					return fmt.Errorf("failed to create csv file: %w", err)
				}
				f, err := os.OpenFile(csvPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return fmt.Errorf("failed to create csv file: %w", err)
				}
				if err := csv.Write(f, allStats, includeReviews); err != nil {
					return fmt.Errorf("failed to create csv file: %w", err)
				}
			}

			fmt.Println()
			return highlights.Write(os.Stdout, allStats, top, includeReviews)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints org-stats version",
		Run: func(cmd *cobra.Command, args []string) {
			if info, ok := debug.ReadBuildInfo(); ok {
				sum := info.Main.Sum
				if sum == "" {
					sum = "none"
				}
				fmt.Printf("https://%s %s @ %s\n", info.Main.Path, info.Main.Version, sum)
			} else {
				fmt.Println("unknown")
			}
		},
	}

	docsCmd := &cobra.Command{
		Use:                   "docs",
		Short:                 "Generates donuts's command line docs",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().DisableAutoGenTag = true
			return doc.GenMarkdownTree(cmd.Root(), "docs")
		},
	}

	rootCmd.Flags().StringVar(&token, "token", "", "github api token (default $GITHUB_TOKEN)")
	rootCmd.MarkFlagRequired(token)

	rootCmd.Flags().StringVarP(&organization, "org", "o", "", "github organization to scan")
	rootCmd.MarkFlagRequired("org")

	rootCmd.Flags().StringSliceVarP(&blacklist, "blacklist", "b", []string{}, "blacklist repos and/or users")
	rootCmd.Flags().IntVar(&top, "top", 3, "how many users to show")
	rootCmd.Flags().StringVar(&githubURL, "github-url", "", "custom github base url (if using github enterprise)")
	rootCmd.Flags().StringVar(&since, "since", "0s", "time to look back to gather info (0s means everything)")
	rootCmd.Flags().BoolVar(&includeReviews, "include-reviews", false, "include pull request reviews in the stats")
	rootCmd.Flags().StringVar(&csvPath, "csv-path", "", "path to write a csv file with all data collected")

	rootCmd.AddCommand(versionCmd, docsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildBlacklists(blacklist []string) ([]string, []string) {
	var userBlacklist []string
	var repoBlacklist []string
	for _, b := range blacklist {
		if strings.HasPrefix(b, "user:") {
			userBlacklist = append(userBlacklist, strings.TrimPrefix(b, "user:"))
		} else if strings.HasPrefix(b, "repo:") {
			repoBlacklist = append(repoBlacklist, strings.TrimPrefix(b, "repo:"))
		} else {
			userBlacklist = append(userBlacklist, b)
			repoBlacklist = append(repoBlacklist, b)
		}
	}
	return userBlacklist, repoBlacklist
}
