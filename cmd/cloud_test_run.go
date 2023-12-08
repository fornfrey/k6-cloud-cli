package cmd

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
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

	showTestSummary := &cobra.Command{
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:   "summary [test-run-id]",
		Short: "Show test run execution summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showTestRunSummary(args[0], client, gs)
		},
	}

	var ofJson bool
	listThresholds := &cobra.Command{
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:   "thresholds [test-run-id]",
		Short: "List test run thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			testRunID := args[0]
			thresholds, err := client.GetCloudTestRunThresholds(testRunID)
			if err != nil {
				return err
			}

			slices.SortFunc(thresholds, func(a, b cloudapi.Threshold) int {
				return strings.Compare(a.Name, b.Name)
			})
			out := NewCloudOutput("", []string{"name", "tainted", "stat", "calculated value"})
			for _, threshold := range thresholds {
				out.Add(map[string]any{
					"name":             threshold.Name,
					"stat":             threshold.Stat,
					"tainted":          threshold.Tainted,
					"calculated value": threshold.CalculatedValue,
				})
			}

			if ofJson {
				out.Json()
			} else {
				out.Print()
			}
			return nil
		},
	}
	listThresholds.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")

	listHttpUrls := &cobra.Command{
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Use:   "httpurls [test-run-id]",
		Short: "List test run HTTP URLs duration stats",
		RunE: func(cmd *cobra.Command, args []string) error {
			testRunID := args[0]
			urls, err := client.GetCloudTestRunHttpUrls(testRunID)
			if err != nil {
				return err
			}
			slices.SortFunc(urls, func(a, b cloudapi.HTTPUrl) int {
				res := strings.Compare(a.Scenario, b.Scenario)
				if res != 0 {
					return res
				}
				res = strings.Compare(a.Name, b.Name)
				if res != 0 {
					return res
				}
				res = strings.Compare(a.Method, b.Method)
				if res != 0 {
					return res
				}
				switch {
				case a.Status < b.Status:
					return -1
				case a.Status > b.Status:
					return 1
				default:
					return 0
				}
			})

			out := NewCloudOutput("", []string{
				"scenario",
				"method",
				"name",
				"status",
				"expected response",
				"count",
				"min",
				"avg",
				"stdev",
				"p(95)",
				"p(99)",
				"max",
			})
			for _, url := range urls {
				out.Add(map[string]any{
					"scenario":          url.Scenario,
					"method":            url.Method,
					"name":              url.Name,
					"status":            url.Status,
					"expected response": url.ExpectedResponse,
					"count":             url.HTTPMetricSummary.RequestsCount,
					"min":               url.HTTPMetricSummary.Duration.Min,
					"avg":               url.HTTPMetricSummary.Duration.Mean,
					"stdev":             url.HTTPMetricSummary.Duration.Stdev,
					"p(95)":             url.HTTPMetricSummary.Duration.P95,
					"p(99)":             url.HTTPMetricSummary.Duration.P99,
					"max":               url.HTTPMetricSummary.Duration.Max,
				})
			}

			if ofJson {
				out.Json()
			} else {
				out.Print()
			}
			return nil
		},
	}
	listHttpUrls.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")

	testrunsSub.AddCommand(
		listTestRun,
		getTestRun,
		downloadTestRun,
		getCloudCmdRunTest(gs),
		showTestSummary,
		listThresholds,
		listHttpUrls,
	)

	return testrunsSub
}

type metricAggregate struct {
	// ui settings
	Label       string
	ValueFormat string
	// whether to use background color
	Background bool

	// query used for fetching the value
	Query string
	Value float64
}
type metricSummary struct {
	Metric     cloudapi.Metric
	Aggregates []*metricAggregate
}

func showTestRunSummary(testRunID string, client *cloudapi.K6CloudClient, gs *state.GlobalState) error {

	var (
		testRun         *cloudapi.CloudTestRun
		testRunSummary  *cloudapi.TestRunSummary
		metricSummaries []metricSummary
		thresholds      []cloudapi.Threshold
		checks          []cloudapi.Check
		httpUrls        []cloudapi.HTTPUrl
	)

	tasks := []chan error{
		runAsync(func() error {
			value, err := client.GetCloudTestRun(testRunID)
			testRun = value
			return err
		}),
		runAsync(func() error {
			value, err := client.GetCloudTestRunSummary(testRunID)
			testRunSummary = value
			return err
		}),
		runAsync(func() error {
			value, err := fetchMetricSummaries(testRunID, client)
			metricSummaries = value
			return err
		}),
		runAsync(func() error {
			value, err := client.GetCloudTestRunThresholds(testRunID)
			thresholds = value
			return err
		}),
		runAsync(func() error {
			value, err := client.GetCloudTestRunChecks(testRunID)
			checks = value
			return err
		}),
		runAsync(func() error {
			value, err := client.GetCloudTestRunHttpUrls(testRunID)
			httpUrls = value
			return err
		}),
	}

	for _, result := range tasks {
		err := <-result
		if err != nil {
			return err
		}
	}

	noColor := gs.Flags.NoColor || !gs.Stdout.IsTTY

	fieldCells := [][]string{
		{"execution", ":"},
		{"duration", ":"},
		{"vuh cost", ":"},
		{"metrics", ":"},
		{"thresholds", ":"},
		{"checks", ":"},
		{"http", ":"},
	}

	rightFieldPadding := strings.Repeat(" ", 2)
	fieldLines := formatTableBlocks(tableBlock{
		cells:   fieldCells,
		padding: 2,
		padchar: ' ',
		flags:   tabwriter.AlignRight,
	})
	for i := range fieldLines {
		fieldLines[i] += rightFieldPadding
	}

	output := new(bytes.Buffer)
	valueColor := getColor(noColor, color.FgCyan)

	execution := "local"
	if len(testRun.Distribution) > 0 {
		execution = "cloud"
	}

	fmt.Fprintf(output, "%s%s\n", fieldLines[0], valueColor.Sprint(execution))
	fmt.Fprintf(output, "%s%s\n", fieldLines[1], valueColor.Sprintf("%.2fs", testRun.ExecutionDuration))
	fmt.Fprintf(output, "%s%s\n\n", fieldLines[2], valueColor.Sprintf("%.2f VUh", testRun.VuhCost))

	currentField := 3
	leftPadding := strings.Repeat(" ", len(fieldLines[currentField])) + rightFieldPadding
	// pritnt metric aggregates
	fmt.Fprint(output, fieldLines[currentField])
	showMetrics(NewLeftPaddedWriter(output, leftPadding), metricSummaries, noColor)
	currentField += 1

	fmt.Fprintln(output)
	fmt.Fprint(output, fieldLines[currentField])
	if err := showThresholds(
		NewLeftPaddedWriter(output, leftPadding),
		testRunSummary,
		thresholds,
		noColor,
	); err != nil {
		return err
	}

	currentField += 1

	fmt.Fprintln(output)
	fmt.Fprint(output, fieldLines[currentField])
	if err := showChecks(
		NewLeftPaddedWriter(output, leftPadding),
		testRunSummary,
		checks,
		noColor,
	); err != nil {
		return err
	}
	currentField += 1

	fmt.Fprintln(output)
	fmt.Fprint(output, fieldLines[currentField])
	if err := showHttpUrls(
		NewLeftPaddedWriter(output, leftPadding),
		testRunSummary,
		httpUrls,
		noColor,
	); err != nil {
		return err
	}

	fmt.Println(output.String())

	return nil
}

func showMetrics(w io.Writer, metricSummaries []metricSummary, noColor bool) error {
	slices.SortFunc(metricSummaries, func(a, b metricSummary) int {
		switch {
		case a.Metric.Origin == b.Metric.Origin:
			return strings.Compare(a.Metric.Name, b.Metric.Name)
		case a.Metric.Origin == "builtin":
			return -1
		case b.Metric.Origin == "builtin":
			return 1
		default:
			return strings.Compare(a.Metric.Origin, b.Metric.Origin)
		}
	})

	numColumns := 0
	for _, m := range metricSummaries {
		if numColumns < len(m.Aggregates) {
			numColumns = len(m.Aggregates)
		}
	}

	labels := make([][]string, len(metricSummaries))
	values := make([][]string, len(metricSummaries))
	valueColor := getColor(noColor, color.FgCyan)
	bgColor := getColor(noColor, color.FgCyan, color.Faint)
	for i, m := range metricSummaries {
		valColumns := make([]string, numColumns)
		for i, agg := range m.Aggregates {
			valuePrint := valueColor.Sprintf
			if agg.Background {
				valuePrint = fmt.Sprintf
			}
			valColumns[i] = valuePrint(agg.ValueFormat, agg.Value)
			if agg.Label != "" {
				valColumns[i] = agg.Label + "=" + valColumns[i]
			}
			if agg.Background {
				valColumns[i] = bgColor.Sprint(valColumns[i])
			}
		}
		// fill the rest of the columns with empty values to preserve formatting
		for i := len(m.Aggregates); i < numColumns; i++ {
			valColumns[i] = ""
		}

		labels[i] = []string{m.Metric.Name, ": "}
		values[i] = valColumns
	}

	lines := formatTableBlocks(tableBlock{
		cells:   labels,
		padding: 3,
		padchar: '.',
	}, tableBlock{
		cells:   values,
		padding: 1,
		padchar: ' ',
	})

	// add a break between builtin and custom metrics
	builtInMetricsEnd := slices.IndexFunc(metricSummaries, func(m metricSummary) bool { return m.Metric.Origin != "builtin" })
	for i, line := range lines {
		if i == builtInMetricsEnd {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, line)
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
					Label:       "avg",
					ValueFormat: "%.2f",
					Query:       "histogram_avg",
				},
				{
					Label:       "min",
					ValueFormat: "%.2f",
					Query:       "histogram_min",
				},
				{
					Label:       "med",
					ValueFormat: "%.2f",
					Query:       "histogram_quantile(0.5)",
				},
				{
					Label:       "max",
					ValueFormat: "%.2f",
					Query:       "histogram_max",
				},
				{
					Label:       "p(90)",
					ValueFormat: "%.2f",
					Query:       "histogram_quantile(0.90)",
				},
				{
					Label:       "p(95)",
					ValueFormat: "%.2f",
					Query:       "histogram_quantile(0.95)",
				},
			},
		}, nil
	case "counter":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					ValueFormat: "%.2f",
					Query:       "increase",
				},
				{
					ValueFormat: "%.2f u/s",
					Query:       "rate",
					Background:  true,
				},
			},
		}, nil
	case "rate":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					ValueFormat: "%.2f%%",
					Query:       "ratio",
				},
				{
					ValueFormat: "✓ %.0f",
					Query:       "increase_nz",
					Background:  true,
				},
				{
					ValueFormat: "✗ %.0f",
					Query:       "increase_z",
					Background:  true,
				},
			},
		}, nil
	case "gauge":
		return metricSummary{
			Metric: metric,
			Aggregates: []*metricAggregate{
				{
					ValueFormat: "%.2f",
					Query:       "avg",
				},
				{
					Label:       "min",
					ValueFormat: "%.2f",
					Query:       "min",
					Background:  true,
				},
				{
					Label:       "max",
					ValueFormat: "%.2f",
					Query:       "max",
					Background:  true,
				},
			},
		}, nil
	default:
		return metricSummary{}, fmt.Errorf("unknown metric type: %s", metric.Type)
	}
}

func fetchMetricSummaries(testRunID string, client *cloudapi.K6CloudClient) ([]metricSummary, error) {
	metrics, err := client.GetCloudTestRunMetrics(testRunID)
	if err != nil {
		return nil, err
	}
	results := make([]chan error, len(metrics))
	metricSummaries := make([]metricSummary, len(metrics))
	for i, metric := range metrics {
		summary, err := buildSummaryForMetric(metric)
		if err != nil {
			return nil, err
		}

		result := make(chan error)
		go func() {
			err := fetchMetricAggregateValues(testRunID, &summary, client)
			result <- err
		}()
		metricSummaries[i] = summary
		results[i] = result
	}

	for _, result := range results {
		err := <-result
		if err != nil {
			return nil, err
		}
	}
	return metricSummaries, nil
}

func fetchMetricAggregateValues(testRunID string, summary *metricSummary, client *cloudapi.K6CloudClient) error {
	results := make([]chan error, len(summary.Aggregates))
	for i := range summary.Aggregates {
		// assing to a new variable to be enclosed by the goroutine
		aggregate := summary.Aggregates[i]
		result := make(chan error)
		go func() {
			queryResult, err := client.GetCloudTestRunMetricsAggregate(
				testRunID, aggregate.Query, summary.Metric.Name, "", "",
			)
			if err != nil {
				result <- err
			} else {
				aggregate.Value = queryResult.Result[0].Values[0][1]
				if aggregate.Query == "ratio" {
					aggregate.Value *= 100
				}
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

func showThresholds(
	w io.Writer,
	testRunSummary *cloudapi.TestRunSummary,
	thresholds []cloudapi.Threshold,
	noColor bool,
) error {
	valueColor := getColor(noColor, color.FgCyan)
	summary := testRunSummary.ThresholdsSummary
	valueColor.Fprintf(w, "%d", summary.Successes)
	fmt.Fprint(w, "/")
	valueColor.Fprintf(w, "%d\n", summary.Total)

	if len(thresholds) == 0 {
		return nil
	}
	slices.SortFunc(thresholds, func(a, b cloudapi.Threshold) int {
		return strings.Compare(a.Name, b.Name)
	})

	labels := make([][]string, len(thresholds))
	values := make([][]string, len(thresholds))
	successColor := getColor(noColor, color.FgGreen)
	failureColor := getColor(noColor, color.FgRed)
	for i, threshold := range thresholds {
		separatorIdx := strings.LastIndex(threshold.Name, ":")
		name := threshold.Name[:separatorIdx]
		condition := strings.ReplaceAll(threshold.Name[separatorIdx+1:], " ", "")
		status := successColor.Sprint("✓")
		if threshold.Tainted {
			status = failureColor.Sprint("✗")
		}

		labels[i] = []string{fmt.Sprintf("%s %s", status, name), ": "}
		values[i] = []string{
			condition,
			threshold.Stat + "=" + valueColor.Sprintf("%.2f", threshold.CalculatedValue),
		}
	}

	lines := formatTableBlocks(tableBlock{
		cells:   labels,
		padding: 3,
		padchar: '.',
	}, tableBlock{
		cells:   values,
		padding: 1,
		padchar: ' ',
	})

	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return nil
}

func showChecks(
	w io.Writer,
	testRunSummary *cloudapi.TestRunSummary,
	checks []cloudapi.Check,
	noColor bool,
) error {
	valueColor := getColor(noColor, color.FgCyan)
	bgColor := getColor(noColor, color.FgCyan, color.Faint)

	summary := testRunSummary.ChecksMetricSummary
	valueColor.Fprintf(w, "%d", summary.HitsSuccesses.Int64)
	fmt.Fprint(w, "/")
	valueColor.Fprintf(w, "%d\n", summary.HitsTotal.Int64)

	if len(checks) == 0 {
		return nil
	}

	slices.SortFunc(checks, func(a, b cloudapi.Check) int {
		return strings.Compare(a.Name, b.Name)
	})

	labels := make([][]string, len(checks))
	values := make([][]string, len(checks))
	for i, check := range checks {
		labels[i] = []string{check.Name, ": "}
		values[i] = []string{
			valueColor.Sprintf("%.2f%%", check.MetricSummary.SuccessRate*100),
			bgColor.Sprintf("✓ %d", check.MetricSummary.SuccessCount),
			bgColor.Sprintf("✗ %d", check.MetricSummary.FailCount),
		}
	}

	lines := formatTableBlocks(tableBlock{
		cells:   labels,
		padding: 0,
		padchar: ' ',
	}, tableBlock{
		cells:   values,
		padding: 1,
		padchar: ' ',
	})

	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return nil
}

func showHttpUrls(
	w io.Writer,
	testRunSummary *cloudapi.TestRunSummary,
	urls []cloudapi.HTTPUrl,
	noColor bool,
) error {

	if len(urls) == 0 {
		return nil
	}

	slices.SortFunc(urls, func(a, b cloudapi.HTTPUrl) int {
		res := strings.Compare(a.Scenario, b.Scenario)
		if res != 0 {
			return res
		}
		res = strings.Compare(a.Name, b.Name)
		if res != 0 {
			return res
		}
		res = strings.Compare(a.Method, b.Method)
		if res != 0 {
			return res
		}
		switch {
		case a.Status < b.Status:
			return -1
		case a.Status > b.Status:
			return 1
		default:
			return 0
		}
	})

	labels := make([][]string, len(urls))
	values := make([][]string, len(urls))
	valueColor := getColor(noColor, color.FgCyan)
	successColor := getColor(noColor, color.FgGreen)
	failureColor := getColor(noColor, color.FgRed)
	for i, url := range urls {
		expected := failureColor.Sprint("✗")
		if url.ExpectedResponse {
			expected = successColor.Sprint("✓")
		}
		labels[i] = []string{fmt.Sprintf("%s %s ", expected, url.Method), valueColor.Sprintf("%d", url.Status), ": "}
		values[i] = []string{
			"count=" + valueColor.Sprintf("%d", url.HTTPMetricSummary.RequestsCount),
			"min=" + valueColor.Sprintf("%.2f", url.HTTPMetricSummary.Duration.Min),
			"avg=" + valueColor.Sprintf("%.2f", url.HTTPMetricSummary.Duration.Mean),
			"stdev=" + valueColor.Sprintf("%.2f", url.HTTPMetricSummary.Duration.Stdev),
			"p(95)=" + valueColor.Sprintf("%.2f", url.HTTPMetricSummary.Duration.P95),
			"p(99)=" + valueColor.Sprintf("%.2f", url.HTTPMetricSummary.Duration.P99),
			"max=" + valueColor.Sprintf("%.2f", url.HTTPMetricSummary.Duration.Max),
		}
	}

	lines := formatTableBlocks(tableBlock{
		cells:   labels,
		padding: 0,
		padchar: ' ',
	}, tableBlock{
		cells:   values,
		padding: 1,
		padchar: ' ',
	})

	bgColor := getColor(noColor, color.FgCyan, color.Faint)
	summary := testRunSummary.HTTPMetricSummary
	summaryLine := strings.Join([]string{
		strings.Join([]string{
			valueColor.Sprintf("%d", summary.RequestsCount-summary.FailuresCount),
			valueColor.Sprintf("%d", summary.RequestsCount),
		}, "/"),
		valueColor.Sprintf("%.2f req/s", summary.RpsMean),
		bgColor.Sprintf("max=%.2f req/s", summary.RpsMax),
	}, "  ")
	fmt.Fprint(w, summaryLine)

	var currentUrl string
	var currentScenario string
	scenarioColor := getColor(noColor, color.FgCyan)
	scenarioWriter := w
	urlWriter := w
	for i := range urls {
		if currentScenario != urls[i].Scenario {
			currentScenario = urls[i].Scenario
			scenarioWriter = NewLeftPaddedWriter(w, "  ")
			fmt.Fprintln(w)
			scenarioColor.Fprintf(w, "%s:", currentScenario)
		}
		if currentUrl != urls[i].Name {
			currentUrl = urls[i].Name
			fmt.Fprintln(scenarioWriter)
			fmt.Fprint(scenarioWriter, currentUrl)
			urlWriter = NewLeftPaddedWriter(scenarioWriter, "  ")
			fmt.Fprintln(urlWriter)
		}
		fmt.Fprintln(urlWriter, lines[i])
	}

	return nil
}

// Concatenate horizontal table blocks applying different formatting per block
// Return an array of resulting lines
func formatTableBlocks(blocks ...tableBlock) []string {
	resultCells := make([][]string, len(blocks[0].cells))
	for i := range resultCells {
		resultCells[i] = make([]string, len(blocks))
	}
	for colIdx, block := range blocks {
		var buffer bytes.Buffer
		w := tabwriter.NewWriter(&buffer, block.minwidth, block.tabwidth, block.padding, block.padchar, block.flags)
		writeToTable(w, block.cells)
		rows := strings.Split(buffer.String(), "\n")
		for rowIdx, row := range rows {
			resultCells[rowIdx][colIdx] = row
		}
	}

	lines := make([]string, len(resultCells))
	for i := range lines {
		lines[i] = strings.Join(resultCells[i], "")
	}
	return lines
}

func writeToTable(writer *tabwriter.Writer, cells [][]string) {
	for i, row := range cells {
		fmt.Fprint(writer, strings.Join(row, "\t"))
		if i != len(cells)-1 {
			fmt.Fprintln(writer)
		}
	}
	writer.Flush()
}

type tableBlock struct {
	cells [][]string

	minwidth int
	tabwidth int
	padding  int
	padchar  byte
	flags    uint
}

// Writer insering padding at the start of each line except the first one
type LeftPaddedWriter struct {
	output      io.Writer
	padBytes    []byte
	isOnNewLine bool
}

func NewLeftPaddedWriter(output io.Writer, padding string) *LeftPaddedWriter {
	return &LeftPaddedWriter{
		output:   output,
		padBytes: []byte(padding),
	}
}

func (w *LeftPaddedWriter) Write(p []byte) (n int, err error) {
	var nTotal int
	for sliceStart := 0; sliceStart < len(p); {
		newlineIdx := bytes.Index(p[sliceStart:], []byte("\n"))

		sliceEnd := newlineIdx + 1
		if newlineIdx == -1 {
			sliceEnd = len(p)
		}

		if len(bytes.TrimSpace(p[sliceStart:sliceEnd])) != 0 && w.isOnNewLine {
			w.isOnNewLine = false
			n, err := w.output.Write(w.padBytes)
			nTotal += n
			if err != nil {
				return nTotal, err
			}
		}

		n, err := w.output.Write(p[sliceStart:sliceEnd])
		nTotal += n
		if err != nil {
			return nTotal, err
		}

		if newlineIdx != -1 {
			w.isOnNewLine = true
		}

		sliceStart = sliceEnd
	}

	return nTotal, nil
}

func runAsync(task func() error) chan error {
	result := make(chan error)
	go func() {
		err := task()
		result <- err
	}()
	return result
}
