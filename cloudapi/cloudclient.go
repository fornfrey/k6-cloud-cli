package cloudapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v3"
)

type Projects struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OrganizationID int       `json:"organization_id"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
	IsDefault      bool      `json:"is_default"`
}

type Organization struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Logo              any       `json:"logo"`
	OwnerID           int       `json:"owner_id"`
	Description       string    `json:"description"`
	BillingAddress    string    `json:"billing_address"`
	BillingCountry    string    `json:"billing_country"`
	BillingEmail      string    `json:"billing_email"`
	VatNumber         string    `json:"vat_number"`
	Created           time.Time `json:"created"`
	Updated           time.Time `json:"updated"`
	IsDefault         bool      `json:"is_default"`
	IsSamlOrg         bool      `json:"is_saml_org"`
	IsGrafanaOrg      bool      `json:"is_grafana_org"`
	GrafanaBillingID  any       `json:"grafana_billing_id"`
	GrafanaOrgName    any       `json:"grafana_org_name"`
	GrafanaOrgSlug    any       `json:"grafana_org_slug"`
	GrafanaStackName  any       `json:"grafana_stack_name"`
	GrafanaStackURL   any       `json:"grafana_stack_url"`
	GrafanaStackID    any       `json:"grafana_stack_id"`
	SubscriptionIds   []int     `json:"subscription_ids"`
	LoadZoneIds       []int     `json:"load_zone_ids"`
	CanTrial          bool      `json:"can_trial"`
	IsPersonal        bool      `json:"is_personal"`
	DataRetentionDays int       `json:"data_retention_days"`
	Vuh               float64   `json:"vuh"`
	VuhMax            float64   `json:"vuh_max"`
	VuhOvercharge     float64   `json:"vuh_overcharge"`
}

type LoadZone struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Vendor       string  `json:"vendor"`
	Country      string  `json:"country"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Configurable bool    `json:"configurable"`
	K6LoadZoneID string  `json:"k6_load_zone_id"`
	Public       bool    `json:"public"`
	Available    bool    `json:"available"`
}

type Account struct {
	User struct {
		ID              int       `json:"id"`
		Email           string    `json:"email"`
		FirstName       string    `json:"first_name"`
		LastName        string    `json:"last_name"`
		Status          int       `json:"status"`
		Country         string    `json:"country"`
		Company         string    `json:"company"`
		Industry        string    `json:"industry"`
		DateJoined      time.Time `json:"date_joined"`
		TimeZone        string    `json:"time_zone"`
		OrganizationIds []int     `json:"organization_ids"`
		GravatarURL     string    `json:"gravatar_url"`
	} `json:"user"`
	Organizations []Organization `json:"organizations"`
	LoadZones     []LoadZone     `json:"load_zones"`
}

type CloudTestRun struct {
	Created           time.Time       `json:"created"`
	Distribution      [][]interface{} `json:"distribution"`
	Duration          int64           `json:"duration"`
	ErrorDetail       string          `json:"error_detail"`
	ExecutionDuration float64         `json:"execution_duration"`
	ID                int64           `json:"id"`
	LoadTime          any             `json:"load_time"`
	Note              string          `json:"note"`
	ProcessingStatus  int             `json:"processing_status"`
	ResultStatus      int             `json:"result_status"`
	RunProcess        string          `json:"run_process"`
	RunStatus         int             `json:"run_status"`
	Started           time.Time       `json:"started"`
	TestID            int64           `json:"test_id"`
	Vus               int             `json:"vus"`
	VuhCost           float64         `json:"vuh_cost"`
	Script            string          `json:"script"`

	RuntimeConfig struct {
		TestRunDetails null.String `json:"testRunDetails"`
	} `json:"k6_runtime_config"`
}

type CloudTest struct {
	Created         time.Time      `json:"created"`
	Updated         time.Time      `json:"updated"`
	CreationProcess string         `json:"creation_process"`
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	ProjectID       int            `json:"project_id"`
	TestRunIds      []int          `json:"test_run_ids"`
	CloudTestRun    []CloudTestRun `json:"test_runs"`
	Script          string         `json:"script"`
}

type ScheduleEnds struct {
	Datetime    string `json:"datetime"`
	Occurrences int64  `json:"occurrences"`
	Type        string `json:"type"`
}

type ScheduleWeekly struct {
	Days []int `json:"days"`
}

type Schedule struct {
	Active      bool           `json:"active"`
	Ends        ScheduleEnds   `json:"ends"`
	Expires     string         `json:"expires"`
	Frequency   string         `json:"string"`
	Id          int64          `json:"id"`
	Interval    int64          `json:"interval"`
	NextRun     string         `json:"next_run"`
	Occurrences int64          `json:"occurrences"`
	Starts      string         `json:"starts"`
	TestId      int64          `json:"test_id"`
	Weekly      ScheduleWeekly `json:"weekly"`
}

type ListSchedulesResponse struct {
	K6Schedules []Schedule `json:"k6-schedules"`
}

type Metric struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Origin    string `json:"origin"`
	TestRunID int    `json:"test_run_id"`
	Type      string `json:"type"`
}

type MetricQueryResult struct {
	ResultType string `json:"resultType"`
	Result     []struct {
		Metric map[string]string `json:"metric"`
		Values [][]float64       `json:"values"`
	} `json:"result"`
}

type Threshold struct {
	CalculatedValue float64 `json:"calculated_value"`
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Stat            string  `json:"stat"`
	Tainted         bool    `json:"tainted"`
}

type StaticIP struct {
	ID                 int    `json:"id"`
	IP                 string `json:"ip"`
	LoadZoneIdentifier string `json:"load_zone_identifier"`
	ProvisioningStatus int    `json:"provisioning_status"`
}

func (s *StaticIP) ProvisioningStatusString() string {
	val, ok := map[int]string{
		0: "Provisioning",
		1: "Provisioning Error",
		5: "Provisioning Error",
		2: "Provisioned",
		3: "Releasing",
		4: "Releasing Error",
		6: "Released",
	}[s.ProvisioningStatus]
	if ok {
		return val
	}
	return "Unknown"
}

type Check struct {
	MetricSummary struct {
		FailCount    int     `json:"fail_count"`
		SuccessCount int     `json:"success_count"`
		SuccessRate  float64 `json:"success_rate"`
	} `json:"metric_summary"`
	Name string `json:"name"`
}

type HTTPUrl struct {
	ExpectedResponse  bool `json:"expected_response"`
	HTTPMetricSummary struct {
		Duration struct {
			Max   float64 `json:"max"`
			Mean  float64 `json:"mean"`
			Min   float64 `json:"min"`
			P95   float64 `json:"p95"`
			P99   float64 `json:"p99"`
			Stdev float64 `json:"stdev"`
		} `json:"duration"`
		RequestsCount int `json:"requests_count"`
	} `json:"http_metric_summary"`
	Method   string `json:"method"`
	Name     string `json:"name"`
	Scenario string `json:"scenario"`
	Status   int    `json:"status"`
}

type TestRunSummary struct {
	ChecksMetricSummary struct {
		HitsSuccesses null.Int `json:"hits_successes"`
		HitsTotal     null.Int `json:"hits_total"`
		Successes     int      `json:"successes"`
		Total         int      `json:"total"`
	} `json:"checks_metric_summary"`
	HTTPMetricSummary struct {
		Duration struct {
			Count int     `json:"count"`
			Max   float64 `json:"max"`
			Mean  float64 `json:"mean"`
			Min   float64 `json:"min"`
			P95   float64 `json:"p95"`
			P99   float64 `json:"p99"`
			Stdev float64 `json:"stdev"`
		} `json:"duration"`
		FailuresCount int     `json:"failures_count"`
		RequestsCount int     `json:"requests_count"`
		RpsMax        float64 `json:"rps_max"`
		RpsMean       float64 `json:"rps_mean"`
	} `json:"http_metric_summary"`
	ThresholdsSummary struct {
		Successes int `json:"successes"`
		Total     int `json:"total"`
	} `json:"thresholds_summary"`
}

func (a *Account) DefaultOrganization() *Organization {
	for _, org := range a.Organizations {
		if org.IsDefault {
			return &org
		}
	}
	return nil
}

// K6CloudClient handles communication with the k6 Cloud API.
type K6CloudClient struct {
	Client
}

func NewK6CloudClient(logger logrus.FieldLogger, token, host, version string, timeout time.Duration) *K6CloudClient {
	return &K6CloudClient{
		Client{
			client:        &http.Client{Timeout: timeout},
			token:         token,
			baseURL:       host,
			version:       version,
			retries:       MaxRetries,
			retryInterval: RetryInterval,
			logger:        logger,
		},
	}
}

func (c *K6CloudClient) GetAccount() (Account, error) {

	url := fmt.Sprintf("%s/v3/account/me", c.baseURL)

	account := Account{}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return account, err
	}
	err = c.Do(req, &account)
	return account, err
}

func (c *K6CloudClient) ListCloudProjects(organizationID string) ([]Projects, error) {
	account, err := c.GetAccount()
	if organizationID == "" {
		organizationID = strconv.Itoa(account.DefaultOrganization().ID)
	}

	url := fmt.Sprintf("%s/v3/organizations/%s/projects", c.baseURL, organizationID)

	projectList := struct {
		Projects []Projects `json:"projects"`
	}{}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	err = c.Do(req, &projectList)
	return projectList.Projects, err
}

func (c *K6CloudClient) ListCloudLoadZones(organizationID string) ([]LoadZone, error) {
	account, err := c.GetAccount()
	if organizationID == "" {
		organizationID = strconv.Itoa(account.DefaultOrganization().ID)
	}
	url := fmt.Sprintf("%s/v3/load-zones?organization_id=%s", c.baseURL, organizationID)

	loadzoneList := struct {
		LoadZones []LoadZone `json:"load_zones"`
	}{}
	req, err := c.NewRequest("GET", url, nil)
	err = c.Do(req, &loadzoneList)
	return loadzoneList.LoadZones, err
}

func (c *K6CloudClient) ListCloudOrganizations() ([]Organization, error) {

	account, err := c.GetAccount()
	return account.Organizations, err
}

func (c *K6CloudClient) ListCloudTests(projectID string) ([]CloudTest, error) {
	if projectID == "" {
		account, err := c.GetAccount()
		if err != nil {
			return nil, err
		}
		organizationID := strconv.Itoa(account.DefaultOrganization().ID)
		projects, err := c.ListCloudProjects(organizationID)
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			if p.IsDefault {
				projectID = strconv.Itoa(p.ID)
			}
		}
	}

	url := fmt.Sprintf("%s/loadtests/v2/tests?$select=id,name,project_id&project_id=%s", c.baseURL, projectID)
	testsList := struct {
		CloudTest []CloudTest `json:"k6-tests"`
	}{}
	req, err := c.NewRequest("GET", url, nil)
	err = c.Do(req, &testsList)
	return testsList.CloudTest, err
}

func (c *K6CloudClient) ListCloudTestRuns(testID string) ([]CloudTestRun, error) {
	url := fmt.Sprintf("%s/loadtests/v2/runs?test_id=%s", c.baseURL, testID)
	testsRunList := struct {
		CloudTestRun []CloudTestRun `json:"k6-runs"`
	}{}
	req, err := c.NewRequest("GET", url, nil)
	err = c.Do(req, &testsRunList)
	return testsRunList.CloudTestRun, err
}

func (c *K6CloudClient) StartCloudTest(testID string) (*CloudTestRun, error) {
	url := fmt.Sprintf("%s/loadtests/v2/tests/%s/start-testrun", c.baseURL, testID)

	req, err := c.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		TestRun CloudTestRun `json:"k6-run"`
	}{}
	err = c.Do(req, &response)
	if err != nil {
		c.logger.Warn(err, url)
		return nil, err
	}
	return &response.TestRun, nil
}

func (c *K6CloudClient) GetCloudTestRun(referenceID string) (*CloudTestRun, error) {
	url := fmt.Sprintf(
		"%s/loadtests/v2/runs/%s?$select=id,duration,script,note,"+
			"result_status,run_status,vuh_cost,distribution,execution_duration,k6_runtime_config",
		c.baseURL,
		referenceID,
	)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		TestRun CloudTestRun `json:"k6-run"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return &response.TestRun, nil
}

func (c *K6CloudClient) GetCloudTest(testId string) (*CloudTest, error) {
	url := fmt.Sprintf("%s/loadtests/v2/tests/%s", c.baseURL, testId)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		CloudTest CloudTest `json:"k6-test"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return &response.CloudTest, nil
}

func (c *K6CloudClient) PatchCloudTest(testId string, data map[string]string) (*CloudTest, error) {
	url := fmt.Sprintf("%s/loadtests/v2/tests/%s", c.baseURL, testId)

	req, err := c.NewRequest("PATCH", url, data)
	if err != nil {
		return nil, err
	}
	response := struct {
		CloudTest CloudTest `json:"k6-test"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return &response.CloudTest, nil
}

func (c *K6CloudClient) ListSchedule(orgId string, jsonOutput bool) ([]Schedule, error) {
	// TODO: can add proj-id support
	url := fmt.Sprintf("%s/v4/schedules?organization_id=%s", c.baseURL, orgId)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	schedules := ListSchedulesResponse{}
	if err := c.Do(req, &schedules); err != nil {
		return nil, err
	}

	return schedules.K6Schedules, err
}

func (c *K6CloudClient) SetSchedule(testId int64, frequency string) error {
	url := fmt.Sprintf("%s/v4/schedules", c.baseURL)

	data := struct {
		TestId    int64          `json:"test_id"`
		Frequency string         `json:"frequency"`
		Ends      ScheduleEnds   `json:"ends"`
		Weekly    ScheduleWeekly `json:"weekly"`
	}{
		testId,
		frequency,
		ScheduleEnds{Type: "never"}, // TODO: possible to allow the schedule to end
		ScheduleWeekly{Days: []int{}},
	}

	req, err := c.NewRequest("POST", url, data)
	if err != nil {
		return err
	}

	if err = c.Do(req, nil); err != nil {
		return err
	}

	fmt.Printf("Successfully created a schedule for test_id %d with a %s cadence", testId, frequency)
	return nil
}

func (c *K6CloudClient) UpdateSchedule(scheduleId int64, frequency string, deactivate bool, activate bool) error {
	url := fmt.Sprintf("%s/v4/schedules/%d", c.baseURL, scheduleId)

	var req *http.Request
	var err error
	var successMsg string

	if deactivate {
		data := struct {
			Active bool `json:"active"`
		}{
			false,
		}

		req, err = c.NewRequest("PATCH", url, data)
		if err != nil {
			return err
		}
		successMsg = fmt.Sprintf("Successfully deactivated schedule %d", scheduleId)
	} else if activate {
		data := struct {
			Active bool `json:"active"`
		}{
			true,
		}

		req, err = c.NewRequest("PATCH", url, data)
		if err != nil {
			return err
		}
		successMsg = fmt.Sprintf("Successfully activated schedule %d", scheduleId)
	} else {
		data := struct {
			Frequency string `json:"frequency"`
		}{
			frequency,
		}

		req, err = c.NewRequest("PATCH", url, data)
		if err != nil {
			return err
		}
		successMsg = fmt.Sprintf("Successfully updated schedule %d with %s frequency", scheduleId, frequency)

	}

	if err = c.Do(req, nil); err != nil {
		return err
	}

	fmt.Println(successMsg)
	return nil
}

func (c *K6CloudClient) DeleteSchedule(scheduleId int64) error {
	url := fmt.Sprintf("%s/v4/schedules/%d", c.baseURL, scheduleId)

	req, err := c.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	if err = c.Do(req, nil); err != nil {
		return err
	}

	fmt.Printf("Successfully deleted schedule %d", scheduleId)
	return nil

}

func (c *K6CloudClient) GetCloudTestRunSummary(referenceID string) (*TestRunSummary, error) {
	url := fmt.Sprintf("%s/loadtests/v4/test_runs(%s)?$select=http_metric_summary,thresholds_summary,checks_metric_summary", c.baseURL, referenceID)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := &TestRunSummary{}

	err = c.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *K6CloudClient) GetCloudTestRunMetrics(referenceID string) ([]Metric, error) {
	url := fmt.Sprintf("%s/cloud/v5/test_runs/%s/metrics", c.baseURL, referenceID)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		Value []Metric `json:"value"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (c *K6CloudClient) GetCloudTestRunMetricsAggregate(
	referenceID, query, metric, start, end string,
) (*MetricQueryResult, error) {
	params := fmt.Sprintf("query='%s',metric='%s'", query, metric)
	if start != "" {
		params += fmt.Sprintf(",start=%s", start)
	}
	if end != "" {
		params += fmt.Sprintf(",end=%s", end)
	}
	url := fmt.Sprintf("%s/cloud/v5/test_runs/%s/query_aggregate_k6(%s)", c.baseURL, referenceID, params)
	url = strings.ReplaceAll(url, " ", "%20")

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		Data MetricQueryResult `json:"data"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *K6CloudClient) GetCloudTestRunThresholds(referenceID string) ([]Threshold, error) {
	url := fmt.Sprintf("%s/loadtests/v4/test_runs(%s)/thresholds?$select=id,name,stat,tainted,calculated_value", c.baseURL, referenceID)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		Value []Threshold `json:"value"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (c *K6CloudClient) GetCloudTestRunChecks(referenceID string) ([]Check, error) {
	url := fmt.Sprintf("%s/loadtests/v4/test_runs(%s)/checks?$select=name,metric_summary&$filter=group_id%%20eq%%20null", c.baseURL, referenceID)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		Value []Check `json:"value"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (c *K6CloudClient) GetCloudTestRunHttpUrls(referenceID string) ([]HTTPUrl, error) {
	url := fmt.Sprintf("%s/loadtests/v4/test_runs(%s)/http_urls?$select=id,scenario_id,group_id,name,method,status,scenario,expected_response,http_metric_summary&$filter=group_id%%20eq%%20null", c.baseURL, referenceID)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response := struct {
		Value []HTTPUrl `json:"value"`
	}{}

	err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (c *K6CloudClient) ListCloudStaticIPs(organizationID string) ([]StaticIP, error) {
	account, err := c.GetAccount()
	if organizationID == "" {
		organizationID = strconv.Itoa(account.DefaultOrganization().ID)
	}

	url := fmt.Sprintf("%s/v4/static-ips/%s", c.baseURL, organizationID)

	list := struct {
		StaticIP []StaticIP `json:"object"`
	}{}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	err = c.Do(req, &list)
	return list.StaticIP, err
}
