package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
	"go.k6.io/k6/cmd/state"
)

func getCloudTestRunCmd(gs *state.GlobalState, client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud testrun
	testrunsSub := &cobra.Command{Use: "testrun"}

	// k6 cloud testrun list
	listTestRun := &cobra.Command{
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
		}}

	getTestRun := &cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "get [test-run-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			test, err := client.GetCloudTestRun(args[0])
			if err != nil {
				return err
			}
			out := NewCloudInfoOutput("%10s", "%v")
			defer out.Print()
			out.Add("ID", test.ID)
			out.Add("Duration", test.Duration)
			out.Add("Note", test.Note)
			out.Add("Script", truncateLines(test.Script, 50, "\n... Use `k6 cloud test download` to the view script"))
			return err
		},
	}

	downloadTestRun := &cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "download [test-run-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			test, err := client.GetCloudTestRun(args[0])
			if err != nil {
				return err
			}
			fmt.Print(test.Script)
			return nil
		},
	}

	testrunsSub.AddCommand(listTestRun, getTestRun, downloadTestRun, getCloudCmdRunTest(gs))

	return testrunsSub
}
