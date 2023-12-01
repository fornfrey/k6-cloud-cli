package cloudapi

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// K6CloudClient handles communication with the k6 Cloud API.
type K6CloudClient struct {
	Client
}

func NewK6CloudClient(logger logrus.FieldLogger, token, host, version string, timeout time.Duration) *K6CloudClient {
	return &K6CloudClient{
		Client{
			client:        &http.Client{Timeout: timeout},
			token:         token,
			baseURL:       "https://api.dev.k6.io",
			version:       version,
			retries:       MaxRetries,
			retryInterval: RetryInterval,
			logger:        logger,
		},
	}
}

func (c *K6CloudClient) ListCloudTests(referenceId string) error {
	url := fmt.Sprintf("%s/v3/organizations/3/projects", c.baseURL)

	req, err := c.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	return c.Do(req, nil)
}
