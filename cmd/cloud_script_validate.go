package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
	"go.k6.io/k6/cmd/state"
	"go.k6.io/k6/lib/consts"
)

func getCloudScriptValidateCmd(gs *state.GlobalState) *cobra.Command {

	c := &cmdCloudRunTest{
		gs:            gs,
		testID:        "",
		showCloudLogs: true,
		exitOnRunning: false,
		uploadOnly:    false,
	}

	exampleText := getExampleText(gs, `
  # Validate a local script.js file.
  {{.}} cloud validate script.js`[1:])

	// k6 cloud validate
	validateSub := &cobra.Command{
		Use:     "validate",
		Example: exampleText,
		Short:   "Validate a k6 script taking your subscriptions into consideration",
		Long:    "Validate a k6 script taking your subscriptions into consideration",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := c.gs.Logger

			currentDiskConf, err := readDiskConfig(c.gs)
			if err != nil {
				return err
			}

			cloudConfig, err := cloudapi.GetConsolidatedConfig(currentDiskConf.Collectors["cloud"], c.gs.Env, "", nil)
			if err != nil {
				return err
			}
			client := cloudapi.NewClient(
				logger, cloudConfig.Token.String, cloudConfig.Host.String, consts.Version, cloudConfig.Timeout.TimeDuration())

			test, err := loadAndConfigureTest(gs, cmd, args, getPartialConfig)
			if err != nil {
				return err
			}

			if err := client.ValidateOptions(test.consolidatedConfig.Options); err != nil {
				return err
			}

			fmt.Println("Script is ok! :)")
			return nil
		}}

	validateSub.Flags().SortFlags = false
	validateSub.Flags().AddFlagSet(c.flagSet())

	return validateSub
}
