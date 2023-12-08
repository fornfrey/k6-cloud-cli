package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudOrganizationCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud organization
	organizationSub := &cobra.Command{Use: "organization", Short: "Manage organizations"}
	// k6 cloud organization list
	listOrganizations := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgs, err := client.ListCloudOrganizations()
			if err != nil {
				return err
			}
			out := NewCloudOutput("%-10v %-25s %-10v\n", []string{"ID", "NAME", "DEFAULT?"})
			defer out.Print()
			for _, org := range orgs {
				out.Add(map[string]any{
					"ID":       org.ID,
					"NAME":     org.Name,
					"DEFAULT?": org.IsDefault,
				})
			}
			return nil
		}}

	organizationSub.AddCommand(listOrganizations)

	return organizationSub
}
