package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudTestRunCmd(client *cloudapi.K6CloudClient) *cobra.Command {
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
			out := NewCloudOutput("%-10v %-10v %-10v %-10v %-30v %-20s \n", []string{"ID", "STATUS", "VUS", "DURATION", "STARTED", "ERROR"})
			defer out.Print()
			for _, t := range tests {
				out.Add(map[string]any{
					"ID":       t.ID,
					"STATUS":   t.RunStatus,
					"VUS":      t.Vus,
					"DURATION": t.Duration,
					"STARTED":  t.Started,
					"ERROR":    t.ErrorDetail,
				})
			}
			return nil
		}})
	return testrunsSub
}
