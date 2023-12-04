package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudLoadZoneCmd(client *cloudapi.K6CloudClient, c *cmdCloud) *cobra.Command {
	// k6 cloud loadzone
	loadzoneSub := &cobra.Command{Use: "loadzone"}
	// k6 cloud loadzone list
	loadzoneSub.AddCommand(&cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			loadzones, err := client.ListCloudLoadZones(c.orgId)
			if err != nil {
				return err
			}
			fs := "%-30v %-25v %-10v %-10v\n"
			fmt.Printf(fs, "NAME", "ID", "CITY", "COUNTRY")
			for _, lz := range loadzones {
				fmt.Printf(fs, lz.Name, lz.K6LoadZoneID, lz.City, lz.Country)
			}
			return nil
		}})
	return loadzoneSub
}
