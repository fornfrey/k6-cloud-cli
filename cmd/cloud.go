package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"

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
		getCloudTestRunCmd(gs, client),
		getCloudScheduleCmd(client),
		getCloudScriptValidateCmd(gs),
	)

	return cmd
}

// CloudOutput will eventually allow us to output JSON and other formats. For now it just helps standardize things.
// TODO: This should be renamed
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

type CloudInfoOutput struct {
	formatHeadings string
	formatInfo     string
	content        [][]any
}

func NewCloudInfoOutput(formatHeading string, formatInfo string) *CloudInfoOutput {
	return &CloudInfoOutput{
		formatHeadings: formatHeading,
		formatInfo:     formatInfo,
	}
}

func (i *CloudInfoOutput) Add(heading string, info any) {
	i.content = append(i.content, []any{heading, info})

}

func (i *CloudInfoOutput) Print() {
	for _, line := range i.content {
		//fmt.Println(strings.Repeat("-", i.longest))
		c, _ := fmt.Printf(i.formatHeadings, line[0])
		fmt.Print(": ")

		infoString, ok := line[1].(string)
		if !ok {
			fmt.Printf(i.formatInfo, line[1])
			fmt.Print("\n")
		} else {
			lines := strings.Split(infoString, "\n")
			for p, l := range lines {
				if p > 0 {
					fmt.Print(strings.Repeat(" ", c), "  ")
				}
				fmt.Printf(i.formatInfo, l)
				fmt.Print("\n")
			}
		}
	}
}

func truncateLines(content string, limit int, message string) string {
	lines := strings.SplitN(content, "\n", limit+1)
	if len(lines) > limit {
		lines[limit] = message
	}
	return strings.Join(lines, "\n")
}
