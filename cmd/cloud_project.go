package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudProjectCmd(client *cloudapi.K6CloudClient, c *cmdCloud) *cobra.Command {
	// k6 cloud project
	projectSub := &cobra.Command{Use: "project"}
	// k6 cloud project list
	projectSub.AddCommand(&cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := client.ListCloudProjects(c.orgId)
			if err != nil {
				return err
			}
			fs := "%-20s %-20s %-10v\n"
			fmt.Printf(fs, "NAME", "DESCRIPTION", "DEFAULT?")
			for _, p := range projects {
				fmt.Printf(fs, p.Name, p.Description, p.IsDefault)
			}
			return nil
		}})
	return projectSub
}
