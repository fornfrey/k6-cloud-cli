package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

// NOTE: This file ends with an underscore so it isn't a test file.

func getCloudTestCmd(client *cloudapi.K6CloudClient, c *cmdCloud) *cobra.Command {
	// k6 cloud test
	testsSub := &cobra.Command{Use: "test"}
	// k6 cloud test list
	testsSub.AddCommand(&cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			tests, err := client.ListCloudTests(c.projId)
			if err != nil {
				return err
			}
			out := NewCloudOutput("%-10v %-25s %-10v \n", []string{"ID", "NAME", "PROJECT ID"})
			defer out.Print()
			for _, t := range tests {
				out.Add(map[string]any{
					"ID":         t.ID,
					"NAME":       t.Name,
					"PROJECT ID": t.ProjectID,
				})
			}
			return nil
		}})
	return testsSub
}
