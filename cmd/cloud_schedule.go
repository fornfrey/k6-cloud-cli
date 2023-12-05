package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudScheduleCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud schedule
	var orgId string
	scheduleSub := &cobra.Command{Use: "schedule"}
	scheduleSub.PersistentFlags().StringVar(&orgId, "org-id", "", "Organization id")

	// k6 cloud schedule list
	listSchedule := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := client.ListSchedule(orgId)
			if err != nil {
				return err
			}
			return nil
		}}

	// k6 cloud schedule set
	setSchedule := &cobra.Command{
		Use:  "set",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			testId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			isValidFrequency := false
			validFrequencies := []string{"never", "hourly", "daily", "weekly", "monthly", "yearly"}
			for _, frequency := range validFrequencies {
				if args[1] == frequency {
					isValidFrequency = true
					break
				}
			}

			if !isValidFrequency {
				errMsg := fmt.Sprintf("%s is not a valid frequency", args[1])
				return errors.New(errMsg)
			}

			fmt.Println(testId)
			err = client.SetSchedule(testId, args[1])
			if err != nil {
				return err
			}
			return nil
		}}

	scheduleSub.AddCommand(listSchedule)
	scheduleSub.AddCommand(setSchedule)

	return scheduleSub
}
