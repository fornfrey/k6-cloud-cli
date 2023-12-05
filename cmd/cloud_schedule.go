package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudScheduleCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud schedule
	var orgId string
	scheduleSub := &cobra.Command{Use: "schedule"}
	scheduleSub.PersistentFlags().StringVar(&orgId, "org-id", "", "Organization id")

	// k6 cloud schedule list
	listSchedules := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := client.ListSchedules(orgId)
			if err != nil {
				return err
			}
			return nil
		}}

	scheduleSub.AddCommand(listSchedules)

	return scheduleSub
}
