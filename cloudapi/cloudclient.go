package cloudapi

import (
	"fmt"
	"net/http"
	"strconv"
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
	VuhMax            int       `json:"vuh_max"`
	VuhOvercharge     int       `json:"vuh_overcharge"`
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
	Created          time.Time `json:"created"`
	Duration         int64     `json:"duration"`
	ErrorDetail      string    `json:"error_detail"`
	ID               int64     `json:"id"`
	LoadTime         any       `json:"load_time"`
	Note             string    `json:"note"`
	ProcessingStatus int       `json:"processing_status"`
	ResultStatus     int       `json:"result_status"`
	RunProcess       string    `json:"run_process"`
	RunStatus        int       `json:"run_status"`
	Started          time.Time `json:"started"`
	TestID           int64     `json:"test_id"`
	Vus              int       `json:"vus"`

	RuntimeConfig struct {
		TestRunDetails null.String `json:"testRunDetails"`
	} `json:"k6_runtime_config"`
}

type CloudTest struct {
	Created         time.Time      `json:"created"`
	CreationProcess string         `json:"creation_process"`
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	ProjectID       int            `json:"project_id"`
	TestRunIds      []int          `json:"test_run_ids"`
	CloudTestRun    []CloudTestRun `json:"test_runs"`
}

type ScheduleEnds struct {
	Datetime    string `json:"datetime"`
	Occurrences int64  `json:"occurrences"`
	Type        string `json:"type"`
}

type ScheduleWeekly struct {
	Days []int `json:"days"` // TODO: what is the correct way to handle in go?
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

func (c *K6CloudClient) StartCloudTest(testID int64) (*CloudTestRun, error) {
	url := fmt.Sprintf("%s/loadtests/v2/tests/%d/start-testrun", c.baseURL, testID)

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
	url := fmt.Sprintf("%s/loadtests/v2/runs/%s?$select=id,duration", c.baseURL, referenceID)

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

func (c *K6CloudClient) ListSchedule(orgId string) error {
	// TODO: can add proj-id support
	url := fmt.Sprintf("%s/v4/schedules?organization_id=%s", c.baseURL, orgId)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	schedules := ListSchedulesResponse{}
	if err := c.Do(req, &schedules); err != nil {
		return err
	}

	// TODO: use common output functionality
	for _, schedule := range schedules.K6Schedules {
		fmt.Println("********** Schedules ***************")
		fmt.Println("test_id", "active", "next_run", "ends_type")
		fmt.Println(schedule.TestId, schedule.Active, schedule.NextRun, schedule.Ends.Type)
	}

	return nil
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

	return c.Do(req, nil)
}

func (c *K6CloudClient) UpdateSchedule(scheduleId int64, frequency string, deactivate bool) error {
	url := fmt.Sprintf("%s/v4/schedules/%d", c.baseURL, scheduleId)

	active := true

	if deactivate {
		active = false
	}

	data := struct {
		Frequency string `json:"frequency"`
		Active    bool   `json:"active"`
	}{
		frequency,
		active,
	}

	req, err := c.NewRequest("PATCH", url, data)
	if err != nil {
		return err
	}

	return c.Do(req, nil)
}

func (c *K6CloudClient) DeleteSchedule(scheduleId int64) error {
	url := fmt.Sprintf("%s/v4/schedules/%d", c.baseURL, scheduleId)

	req, err := c.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	return c.Do(req, nil)
}
