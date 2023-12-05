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
			return client.ListSchedule(orgId)
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

			return client.SetSchedule(testId, args[1])
		}}

	// k6 cloud schedule update
	var deactivate bool
	updateSchedule := &cobra.Command{
		Use:  "update",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scheduleId, err := strconv.ParseInt(args[0], 10, 64)
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

			return client.UpdateSchedule(scheduleId, args[1], deactivate)
		}}

	updateSchedule.Flags().BoolVar(&deactivate, "deactivate", false, "Deactivate the schedule")

	// k6 cloud schedule delete
	deleteSchedule := &cobra.Command{
		Use:  "delete",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scheduleId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			return client.DeleteSchedule(scheduleId)
		}}

	scheduleSub.AddCommand(listSchedule)
	scheduleSub.AddCommand(setSchedule)
	scheduleSub.AddCommand(updateSchedule)
	scheduleSub.AddCommand(deleteSchedule)

	return scheduleSub
}
