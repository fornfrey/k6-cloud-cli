package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.k6.io/k6/cloudapi"
	"go.k6.io/k6/cmd/state"
	"go.k6.io/k6/lib/consts"
)

func getCloudClient(gs *state.GlobalState) (*cloudapi.K6CloudClient, error) {
	currentDiskConf, err := readDiskConfig(gs)
	if err != nil {
		return nil, err
	}

	cloudConfig, err := cloudapi.GetConsolidatedConfig(currentDiskConf.Collectors["cloud"], gs.Env, "", nil)
	if err != nil {
		return nil, err
	}

	if !cloudConfig.Token.Valid {
		return nil, errors.New("Not logged in, please use `k6 login cloud`.") //nolint:golint,revive,stylecheck
	}

	apiClient := cloudapi.NewK6CloudClient(
		gs.Logger, cloudConfig.Token.String, cloudConfig.APIHost.String, consts.Version, cloudConfig.Timeout.TimeDuration())
	return apiClient, nil
}

func getCmdCloud(gs *state.GlobalState) *cobra.Command {
	client, err := getCloudClient(gs)
	if err != nil {
		gs.Logger.Error(err)
		os.Exit(1)
	}
	cmd := &cobra.Command{Use: "cloud"}

	cmd.AddCommand(
		getCloudProjectCmd(client),
		getCloudLoadZoneCmd(client),
		getCloudOrganizationCmd(client),
		getCloudTestCmd(gs, client),
		getCloudTestRunCmd(client),
		getCloudScheduleCmd(client),
	)

	return cmd
}

// CloudOutput will eventually allow us to putput JSON and other formats. For now it just helps standarise things.
type CloudOutput struct {
	format   string
	headings []string
	content  []map[string]any
}

func NewCloudOutput(format string, headings []string) *CloudOutput {
	return &CloudOutput{format: format, headings: headings}
}

func (o *CloudOutput) Add(line map[string]any) {
	o.content = append(o.content, line)
}

func (o *CloudOutput) PrintHeading() {
	h := make([]interface{}, len(o.headings))
	for i := range o.headings {
		h[i] = o.headings[i]
	}
	fmt.Printf(o.format, h...)
}

func (o *CloudOutput) PrintLine(line map[string]any) {
	var l []any
	for _, heading := range o.headings {
		l = append(l, line[heading])
	}
	fmt.Printf(o.format, l...)
}

func (o *CloudOutput) Print() {
	o.PrintHeading()
	for _, line := range o.content {
		o.PrintLine(line)
	}
}
