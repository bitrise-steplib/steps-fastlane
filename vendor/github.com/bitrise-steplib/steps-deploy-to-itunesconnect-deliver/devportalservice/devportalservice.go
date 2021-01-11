package devportalservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/bitrise-io/go-utils/log"
)

const (
	bitriseBuildURLKey      = "BITRISE_BUILD_URL"
	bitriseBuildAPITokenKey = "BITRISE_BUILD_API_TOKEN"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AppleDeveloperConnectionProvider ...
type AppleDeveloperConnectionProvider interface {
	GetAppleDeveloperConnection(buildURL, buildAPIToken string) (*AppleDeveloperConnection, error)
}

// BitriseClient ...
type BitriseClient struct {
	httpClient httpClient
}

// NewBitriseClient ...
func NewBitriseClient(client httpClient) *BitriseClient {
	return &BitriseClient{
		httpClient: client,
	}
}

const appleDeveloperConnectionPath = "apple_developer_portal_data.json"

// GetAppleDeveloperConnection ...
func (c *BitriseClient) GetAppleDeveloperConnection(buildURL, buildAPIToken string) (*AppleDeveloperConnection, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", buildURL, appleDeveloperConnectionPath), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("BUILD_API_TOKEN", buildAPIToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// On error, any Response can be ignored
		return nil, fmt.Errorf("failed to perform request, error: %s", err)
	}

	// The client must close the response body when finished with it
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Warnf("Failed to close response body, error: %s", cerr)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body, error: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NetworkError{Status: resp.StatusCode, Body: string(body)}
	}

	var connection AppleDeveloperConnection
	if err := json.Unmarshal([]byte(body), &connection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response (%s), error: %s", body, err)
	}

	return &connection, nil
}

// cookie ...
type cookie struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Value     string `json:"value"`
	Domain    string `json:"domain"`
	Secure    bool   `json:"secure"`
	Expires   string `json:"expires,omitempty"`
	MaxAge    int    `json:"max_age,omitempty"`
	Httponly  bool   `json:"httponly"`
	ForDomain *bool  `json:"for_domain,omitempty"`
}

// AppleDeveloperConnection ...
type AppleDeveloperConnection struct {
	AppleID              string              `json:"apple_id"`
	Password             string              `json:"password"`
	ConnectionExpiryDate string              `json:"connection_expiry_date"`
	SessionCookies       map[string][]cookie `json:"session_cookies"`
}

// Expiry ...
func (c *AppleDeveloperConnection) Expiry() *time.Time {
	t, err := time.Parse(time.RFC3339, c.ConnectionExpiryDate)
	if err != nil {
		return nil
	}
	return &t
}

// Expired ...
func (c *AppleDeveloperConnection) Expired() bool {
	expiry := c.Expiry()
	if expiry == nil {
		return false
	}
	return expiry.Before(time.Now())
}

// TFASession ...
func (c *AppleDeveloperConnection) TFASession() (string, error) {
	var rubyCookies []string
	for _, cookie := range c.SessionCookies["https://idmsa.apple.com"] {
		if rubyCookies == nil {
			rubyCookies = append(rubyCookies, "---"+"\n")
		}

		if cookie.ForDomain == nil {
			b := true
			cookie.ForDomain = &b
		}

		tmpl, err := template.New("").Parse(`- !ruby/object:HTTP::Cookie
  name: {{.Name}}
  value: {{.Value}}
  domain: {{.Domain}}
  for_domain: {{.ForDomain}}
  path: "{{.Path}}"
`)
		if err != nil {
			return "", fmt.Errorf("failed to parse template: %s", err)
		}

		var b bytes.Buffer
		err = tmpl.Execute(&b, cookie)
		if err != nil {
			return "", fmt.Errorf("failed to execute template on cookie: %s: %s", cookie.Name, err)
		}

		rubyCookies = append(rubyCookies, b.String()+"\n")
	}
	return strings.Join(rubyCookies, ""), nil
}
