package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
	"go.k6.io/k6/cmd/state"
)

func getCloudScheduleCmd(gs *state.GlobalState, client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud schedule
	var orgId string
	var jsonOutput bool

	exampleText := getExampleText(gs, `
  # List all schedules for an organization.
  {{.}} cloud schedule list --org-id ID

  # Create a schedule for a test.
  {{.}} cloud schedule set TEST_ID FREQUENCY

  # Update a schedule.
  {{.}} cloud schedule update SCHEDULE_ID FREQUENCY
  {{.}} cloud schedule update SCHEDULE_ID --deactivate
  {{.}} cloud schedule update SCHEDULE_ID --activate

  # Delete a schedule.
  {{.}} cloud schedule delete SCHEDULE_ID`[1:])

	scheduleSub := &cobra.Command{
		Use:     "schedule",
		Example: exampleText,
		Short:   "Configure test schedules",
		Long:    "Configure test schedules",
	}
	scheduleSub.PersistentFlags().StringVar(&orgId, "org-id", "", "Organization id")

	// k6 cloud schedule list
	listSchedule := &cobra.Command{
		Use:   "list",
		Short: "List all test schedules for an organization",
		Long:  "List all test schedules for an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			schedules, err := client.ListSchedule(orgId, jsonOutput)
			if err != nil {
				return err
			}

			out := NewTabbedCloudOutput([]string{"%d", "%d", "%t", "%s", "%s"}, []string{"schedule_id", "test_id", "active", "next_run", "ends_type"})
			if jsonOutput {
				defer out.Json()
			} else {
				defer out.PrintTabled()
			}
			for _, schedule := range schedules {
				out.Add(map[string]any{
					"schedule_id": schedule.Id,
					"test_id":     schedule.TestId,
					"active":      schedule.Active,
					"next_run":    schedule.NextRun,
					"ends_type":   schedule.Ends.Type,
				})
			}

			return nil
		}}

	listSchedule.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON")

	// k6 cloud schedule set
	setSchedule := &cobra.Command{
		Use:   "set",
		Short: "Create a new schedule",
		Long:  "Create a new schedule",
		Args:  cobra.MinimumNArgs(2),
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
	var activate bool
	updateSchedule := &cobra.Command{
		Use:   "update",
		Short: "Update a schedule",
		Long:  "Update a schedule",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scheduleId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			frequency := ""

			if len(args) > 1 {

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

				frequency = args[1]
			}

			return client.UpdateSchedule(scheduleId, frequency, deactivate, activate)

		}}

	updateSchedule.Flags().BoolVar(&deactivate, "deactivate", false, "Deactivate the schedule")
	updateSchedule.Flags().BoolVar(&activate, "activate", false, "Activate the schedule")

	// k6 cloud schedule delete
	deleteSchedule := &cobra.Command{
		Use:   "delete",
		Short: "Delete a schedule",
		Long:  "Delete a schedule",
		Args:  cobra.MinimumNArgs(1),
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
