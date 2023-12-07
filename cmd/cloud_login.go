package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cmd/state"
)

func getCloudLoginCmd(gs *state.GlobalState) *cobra.Command {
	// k6 cloud login
	cmdLoginCloud := getCmdLoginCloud(gs)

	exampleText := getExampleText(gs, `
  # Show the stored token.
  {{.}} cloud login -s

  # Store a token.
  {{.}} cloud login -t YOUR_TOKEN

  # Log in with an email/password.
  {{.}} cloud login`[1:])

	loginCloudCommand := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with k6 Cloud",
		Long: `Authenticate with k6 Cloud",

This will set the default token used when just "k6 run -o cloud" is passed.`,
		Example: exampleText,
		Args:    cobra.NoArgs,
		RunE:    cmdLoginCloud.RunE,
	}

	loginCloudCommand.Flags().StringP("token", "t", "", "specify `token` to use")
	loginCloudCommand.Flags().BoolP("show", "s", false, "display saved token and exit")
	loginCloudCommand.Flags().BoolP("reset", "r", false, "reset token")

	return loginCloudCommand
}
