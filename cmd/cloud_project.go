package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudProjectCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud project
	projectSub := &cobra.Command{Use: "project"}
	// k6 cloud project list
	var orgId string
	listProjects := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := client.ListCloudProjects(orgId)
			if err != nil {
				return err
			}
			out := NewCloudOutput("%-20s %-20s %-10v\n", []string{"NAME", "DESCRIPTION", "DEFAULT?"})
			defer out.Print()
			for _, p := range projects {
				out.Add(map[string]any{
					"NAME":        p.Name,
					"DESCRIPTION": p.Description,
					"DEFAULT?":    p.IsDefault,
				})
			}
			return nil
		}}
	listProjects.Flags().StringVar(&orgId, "org-id", "", "Organization id")

	projectSub.AddCommand(listProjects)

	return projectSub
}
