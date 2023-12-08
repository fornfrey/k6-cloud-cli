package cmd

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
)

func getCloudMetricsCmd(client *cloudapi.K6CloudClient) *cobra.Command {
	metricsSub := &cobra.Command{Use: "metrics"}

	var testRunID string
	var ofJson bool
	listMetrics := &cobra.Command{
		Use:   "list",
		Short: "List metrics in the test run",
		RunE: func(cmd *cobra.Command, args []string) error {
			metrics, err := client.GetCloudTestRunMetrics(testRunID)
			if err != nil {
				return err
			}
			slices.SortFunc(metrics, func(a, b cloudapi.Metric) int {
				switch {
				case a.Origin == b.Origin:
					return strings.Compare(a.Name, b.Name)
				case a.Origin == "builtin":
					return -1
				case b.Origin == "builtin":
					return 1
				default:
					return strings.Compare(a.Origin, b.Origin)
				}
			})
			out := NewCloudOutput("", []string{"Name", "Type", "Origin"})
			for _, metric := range metrics {
				out.Add(map[string]any{
					"Name":   metric.Name,
					"Type":   metric.Type,
					"Origin": metric.Origin,
				})
			}

			if ofJson {
				out.Json()
			} else {
				out.Print()
			}

			return nil
		}}
	listMetrics.Flags().StringVarP(&testRunID, "test-run-id", "t", "", "Load Test Run ID")
	listMetrics.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")
	listMetrics.MarkFlagRequired("test-run-id")

	var (
		metric string
		query  string
		start  string
		end    string
	)
	queryMetric := &cobra.Command{
		Use:   "aggregate",
		Short: "Query metric aggregate values",
		RunE: func(cmd *cobra.Command, args []string) error {
			queryResult, err := client.GetCloudTestRunMetricsAggregate(testRunID, query, metric, start, end)
			if err != nil {
				return err
			}

			if ofJson {
				bytes, err := json.Marshal(queryResult)
				if err != nil {
					return err
				}
				fmt.Print(string(bytes))
				return nil
			}

			formatMetricLabels := func(labels map[string]string) string {
				kvPairs := make([]string, 0, len(labels))
				for k, v := range labels {
					if k != "__name__" {
						kvPairs = append(kvPairs, fmt.Sprintf("%s=\"%s\"", k, v))
					}
				}
				slices.Sort(kvPairs)
				return fmt.Sprintf("{%s}", strings.Join(kvPairs, ","))
			}

			rows := make([][]string, len(queryResult.Result))
			for i, result := range queryResult.Result {
				rows[i] = []string{formatMetricLabels(result.Metric), fmt.Sprint(result.Values[0][1])}
			}
			slices.SortFunc(rows, func(a, b []string) int {
				return strings.Compare(a[0], b[0])
			})

			out := NewCloudOutput("", []string{"labels", "value"})
			for _, row := range rows {
				out.Add(map[string]any{
					"labels": row[0],
					"value":  row[1],
				})
			}

			out.Print()

			return nil
		}}
	queryMetric.Flags().StringVarP(&testRunID, "test-run-id", "t", "", "Load Test Run ID")
	queryMetric.Flags().StringVarP(&metric, "metric", "m", "",
		"Required: name of the metric to query, with optional labels selector. Example: `http_reqs{expected_response=“true”}")
	queryMetric.Flags().StringVar(&query, "query", "", "Required: query expression for selecting metrics")
	queryMetric.Flags().StringVarP(&start, "start", "s", "",
		"The start timestamp for the query range. Default is test run start. Example: '2022-01-02T10:03:00Z'")
	queryMetric.Flags().StringVarP(&end, "end", "e", "",
		"The end timestamp for the query range. Default is test run end. Example: '2022-01-02T10:13:00Z'")
	queryMetric.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")
	queryMetric.MarkFlagRequired("test-run-id")
	queryMetric.MarkFlagRequired("metric")
	queryMetric.MarkFlagRequired("query")

	metricsSub.AddCommand(listMetrics, queryMetric)

	return metricsSub

}
