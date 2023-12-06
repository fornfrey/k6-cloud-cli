package cmd

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.k6.io/k6/cloudapi"
	"go.k6.io/k6/cmd/state"
)

type cmdShowTestSummary struct {
	client    *cloudapi.K6CloudClient
	testRunID string

	thresholds bool
	httpUrls   bool
	checks     bool
}

func (c *cmdShowTestSummary) shouldShowAll() bool {
	return !c.thresholds && !c.httpUrls && !c.checks
}

func getCloudTestRunCmd(gs *state.GlobalState, client *cloudapi.K6CloudClient) *cobra.Command {
	// k6 cloud testrun
	testrunsSub := &cobra.Command{Use: "testrun"}

	// k6 cloud testrun list
	listTestRun := &cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "list [test-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			tests, err := client.ListCloudTestRuns(args[0])
			if err != nil {
				return err
			}
			out := NewCloudOutput("%-10v %-10v %-10v %-10v %-30v %-20s \n", []string{"ID", "STATUS", "VUS", "DURATION", "STARTED", "ERROR"})
			defer out.Print()
			for _, t := range tests {
				out.Add(map[string]any{
					"ID":       t.ID,
					"STATUS":   t.RunStatus,
					"VUS":      t.Vus,
					"DURATION": t.Duration,
					"STARTED":  t.Started,
					"ERROR":    t.ErrorDetail,
				})
			}
			return nil
		}}

	getTestRun := &cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "get [test-run-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			test, err := client.GetCloudTestRun(args[0])
			if err != nil {
				return err
			}
			out := NewCloudInfoOutput("%10s", "%v")
			defer out.Print()
			out.Add("ID", test.ID)
			out.Add("Duration", test.Duration)
			out.Add("Note", test.Note)
			out.Add("Script", truncateLines(test.Script, 50, "\n... Use `k6 cloud test download` to the view script"))
			return err
		},
	}

	downloadTestRun := &cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "download [test-run-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			test, err := client.GetCloudTestRun(args[0])
			if err != nil {
				return err
			}
			fmt.Print(test.Script)
			return nil
		},
	}

	cmdShowTestSummary := cmdShowTestSummary{}
	showTestSummary := &cobra.Command{
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:  "summary [test-run-id]",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdShowTestSummary.testRunID = args[0]
			return cmdShowTestSummary.showTestRunSummary(client, gs)
		},
	}
	showTestSummary.Flags().BoolVar(&cmdShowTestSummary.thresholds, "thresholds", false, "show thresholds")
	showTestSummary.Flags().BoolVar(&cmdShowTestSummary.httpUrls, "http-urls", false, "show http urls")

	testrunsSub.AddCommand(listTestRun, getTestRun, downloadTestRun, getCloudCmdRunTest(gs), showTestSummary)

	return testrunsSub
}

type metricAggregate struct {
	// label diplayed in the ui
	Label string
	// query used for fetching the value
	Query string
	Value float64
}
type metricSummary struct {
	Metric     cloudapi.Metric
	Aggregates []*metricAggregate
}

func (c *cmdShowTestSummary) showTestRunSummary(client *cloudapi.K6CloudClient, gs *state.GlobalState) error {

	metrics, err := client.GetCloudTestRunMetrics(c.testRunID)
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
	results := make([]chan error, len(metrics))
	metricSummaries := make([]metricSummary, len(metrics))
	for i, metric := range metrics {
		summary, err := buildSummaryForMetric(metric)
		if err != nil {
			return err
		}

		result := make(chan error)
		go func() {
			err := fetchMetricAggregateValues(c.testRunID, &summary, client)
			result <- err
		}()
		metricSummaries[i] = summary
		results[i] = result
	}

	for _, result := range results {
		err := <-result
		if err != nil {
			return err
		}
	}

	// Writing and flushing labels and values to different buffers first as they are formatted differently
	var labelsBuf bytes.Buffer
	labelsWriter := tabwriter.NewWriter(&labelsBuf, 0, 0, 3, '.', 0)
	for _, summary := range metricSummaries {
		if _, err := fmt.Fprintf(labelsWriter, "%s\t:\n", summary.Metric.Name); err != nil {
			return err
		}
	}
	if err := labelsWriter.Flush(); err != nil {
		return err
	}
	labelRows := bytes.Split(labelsBuf.Bytes(), []byte("\n"))

	var valuesBuf bytes.Buffer
	valuesWriter := tabwriter.NewWriter(&valuesBuf, 0, 1, 1, ' ', 0)
	for _, summary := range metricSummaries {
		columns := make([]string, len(summary.Aggregates))
		for i, agg := range summary.Aggregates {
			columns[i] = fmt.Sprintf(agg.Label, agg.Value)
		}
		row := strings.Join(columns, "\t")
		if _, err := fmt.Fprintln(valuesWriter, row); err != nil {
			return err
		}
	}
	if err := valuesWriter.Flush(); err != nil {
		return err
	}
	valueRows := bytes.Split(valuesBuf.Bytes(), []byte("\n"))

	for i, label := range labelRows {
		fmt.Printf("%s %s\n", label, valueRows[i])
	}

	return nil
}

func buildSummaryForMetric(metric cloudapi.Metric) (metricSummary, error) {
	switch metric.Type {
	case "trend":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					Label: "avg=%.2f",
					Query: "histogram_avg",
				},
				{
					Label: "min=%.2f",
					Query: "histogram_min",
				},
				{
					Label: "med=%.2f",
					Query: "histogram_quantile(0.5)",
				},
				{
					Label: "max=%.2f",
					Query: "histogram_max",
				},
				{
					Label: "p(90)=%.2f",
					Query: "histogram_quantile(0.90)",
				},
				{
					Label: "p(95)=%.2f",
					Query: "histogram_quantile(0.95)",
				},
			},
		}, nil
	case "counter":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					Label: "%.2f",
					Query: "increase",
				},
				{
					Label: "%.2f u/s",
					Query: "rate",
				},
			},
		}, nil
	case "rate":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					Label: "%.2f%%",
					Query: "ratio",
				},
				{
					Label: "✓ %.2f",
					Query: "increase_nz",
				},
				{
					Label: "✗ %.2f u/s",
					Query: "increase_z",
				},
			},
		}, nil
	case "gauge":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					Label: "%.2f",
					Query: "avg",
				},
				{
					Label: "min=%.2f",
					Query: "min",
				},
				{
					Label: "max=%.2f",
					Query: "max",
				},
			},
		}, nil
	default:
		return metricSummary{}, fmt.Errorf("unknown metric type: %s", metric.Type)
	}
}

func fetchMetricAggregateValues(testRunID string, summary *metricSummary, client *cloudapi.K6CloudClient) error {
	results := make([]chan error, len(summary.Aggregates))
	for i := range summary.Aggregates {
		// assing to a new variable to be enclosed by the goroutine
		aggregate := summary.Aggregates[i]
		result := make(chan error)
		go func() {
			value, err := client.GetCloudTestRunMetricsAggregate(testRunID, aggregate.Query, summary.Metric.Name)
			if err != nil {
				result <- err
			} else {
				aggregate.Value = value
				result <- nil
			}
		}()
		results[i] = result
	}
	for _, result := range results {
		err := <-result
		if err != nil {
			return err
		}
	}
	return nil
}
