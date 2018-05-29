package main

import (
	"fmt"
	"log"
	"os"

	"github.com/educlos/testrail"
	"github.com/urfave/cli"

	"github.com/whoshuu/trailer/spec"
)

func main() {
	username := os.Getenv("TESTRAIL_USERNAME")
	token := os.Getenv("TESTRAIL_TOKEN")

	if username == "" || token == "" {
		log.Fatalf("Need to set TESTRAIL_USERNAME and TESTRAIL_TOKEN")
	}

	var (
		verbose bool
		dry     bool
		runID   int
		comment string
	)

	app := cli.NewApp()
	app.HideHelp = true
	app.HideVersion = true
	app.Name = "trailer"
	app.Usage = "TestRail command line utility"
	app.Commands = []cli.Command{
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "Upload JUnit XML reports to TestRail",
			Flags: []cli.Flag{
				// TODO: Respect verbosity and use a proper logging library
				cli.BoolFlag{
					Name:        "verbose, v",
					Usage:       "turn on debug logs",
					Destination: &verbose,
				},
				cli.BoolFlag{
					Name:        "dry, d",
					Usage:       "print readable results without updating TestRail run",
					Destination: &dry,
				},
				cli.IntFlag{
					Name:        "run-id, r",
					Usage:       "TestRail run ID to target for the update",
					Destination: &runID,
				},
				cli.StringFlag{
					Name:        "comment, c",
					Usage:       "prefix to use when commenting on TestRail updates",
					Destination: &comment,
				},
			},
			ArgsUsage: "[input *.xml files...]",
			Action: func(c *cli.Context) error {
				if runID == 0 {
					log.Fatalf("Must set --run-id to a non-zero integer")
				}

				updates := spec.Updates{
					ResultMap: map[int]spec.Update{},
				}

				suites := spec.JUnitTestSuites{}
				for _, file := range c.Args() {
					newSuites, err := spec.ParseFile(file)
					if err != nil {
						log.Fatalf(fmt.Sprintf("Failed to parse file: %s", err))
					}

					suites.Suites = append(suites.Suites, newSuites...)
				}

				updates.AddSuites(comment, suites)

				results, err := updates.CreatePayload()
				if err != nil {
					log.Fatalf(fmt.Sprintf("Failed to create results payload: %s", err))
				}

				if !dry {
					client := testrail.NewClient("https://docker.testrail.com", username, token)
					r, err := client.AddResultsForCases(runID, results)
					if err != nil {
						log.Fatalf(fmt.Sprintf("Failed to upload test results to TestRail: %s", err))
					}

					for _, res := range r {
						fmt.Printf("%+v\n", res)
					}
				}

				return nil
			},
		},
	}

	app.Run(os.Args)
}
