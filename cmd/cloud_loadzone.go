package cmd

import (
	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudLoadZoneCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud loadzone
	loadzoneSub := &cobra.Command{Use: "loadzone"}
	// k6 cloud loadzone list
	var orgId string
	listLoadZones := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			loadzones, err := client.ListCloudLoadZones(orgId)
			if err != nil {
				return err
			}

			out := NewCloudOutput("%-30v %-25v %-10v %-10v %-15v %-15v\n", []string{"Name", "ID", "City", "Country", "Latitude", "Longitude"})
			defer out.Print()
			for _, lz := range loadzones {
				out.Add(map[string]any{
					"Name":      lz.Name,
					"ID":        lz.K6LoadZoneID,
					"City":      lz.City,
					"Country":   lz.Country,
					"Latitude":  lz.Latitude,
					"Longitude": lz.Longitude,
				})
			}
			return nil
		}}
	listLoadZones.Flags().StringVar(&orgId, "org-id", "", "Organization id")

	loadzoneSub.AddCommand(listLoadZones)

	return loadzoneSub
}
