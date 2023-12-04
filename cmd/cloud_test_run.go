package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudTestRunCmd(client *cloudapi.K6CloudClient, c *cmdCloud) *cobra.Command {
	// k6 cloud testrun
	testrunsSub := &cobra.Command{Use: "testrun"}

	// k6 cloud testrun list
	testrunsSub.AddCommand(&cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "list [test-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			tests, err := client.ListCloudTestRuns(args[0])
			if err != nil {
				return err
			}
			fs := "%-10v %-10v %-10v %-10v %-30v %-20s \n"
			fmt.Printf(fs, "ID", "Status", "VUs", "Duration", "Started", "ERROR")
			for _, t := range tests {
				fmt.Printf(fs, t.ID, t.RunStatus, t.Vus, t.Duration, t.Started, t.ErrorDetail)
			}
			return nil
		}})
	return testrunsSub
}
