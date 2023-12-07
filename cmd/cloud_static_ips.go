package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudStaticIPCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud staticip
	staticIPSub := &cobra.Command{Use: "staticip"}
	var ofJson bool
	// k6 cloud staticip list
	var orgId string
	listStaticIP := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			staticIPs, err := client.ListCloudStaticIPs(orgId)
			if err != nil {
				return err
			}
			out := NewCloudOutput("%-10v %-20s %-20s %-10v\n", []string{"ID", "IP", "Load Zone", "Status"})
			if ofJson {
				defer out.Json()
			} else {
				defer out.Print()
			}
			for _, s := range staticIPs {
				out.Add(map[string]any{
					"ID":        s.ID,
					"IP":        s.IP,
					"Load Zone": s.LoadZoneIdentifier,
					"Status":    s.ProvisioningStatusString(),
				})
			}
			return nil
		}}
	listStaticIP.Flags().StringVar(&orgId, "org-id", "", "Organization id")
	listStaticIP.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")

	staticIPSub.AddCommand(listStaticIP)
	return staticIPSub
}
