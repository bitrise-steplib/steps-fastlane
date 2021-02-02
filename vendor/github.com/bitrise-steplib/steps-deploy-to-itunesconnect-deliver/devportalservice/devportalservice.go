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

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AppleDeveloperConnectionProvider ...
type AppleDeveloperConnectionProvider interface {
	GetAppleDeveloperConnection() (*AppleDeveloperConnection, error)
}

// BitriseClient implements AppleDeveloperConnectionProvider through the Bitrise.io API.
type BitriseClient struct {
	httpClient              httpClient
	buildURL, buildAPIToken string

	readBytesFromFile func(pth string) ([]byte, error)
}

// NewBitriseClient creates a new instance of BitriseClient.
func NewBitriseClient(client httpClient, buildURL, buildAPIToken string) *BitriseClient {
	return &BitriseClient{
		httpClient:        client,
		buildURL:          buildURL,
		buildAPIToken:     buildAPIToken,
		readBytesFromFile: fileutil.ReadBytesFromFile,
	}
}

const appleDeveloperConnectionPath = "apple_developer_portal_data.json"

// GetAppleDeveloperConnection fetches the Bitrise.io session-based Apple Developer connection.
func (c *BitriseClient) GetAppleDeveloperConnection() (*AppleDeveloperConnection, error) {
	var rawCreds []byte
	var err error

	if strings.HasPrefix(c.buildURL, "file://") {
		rawCreds, err = c.readBytesFromFile(strings.TrimPrefix(c.buildURL, "file://"))
	} else {
		rawCreds, err = c.download()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch authentication credentials: %v", err)
	}

	type data struct {
		*SessionConnection
		*JWTConnection
		TestDevices []TestDevice `json:"test_devices"`
	}
	var d data
	if err := json.Unmarshal([]byte(rawCreds), &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authentication credentials from response (%s): %s", rawCreds, err)
	}

	if d.JWTConnection != nil {
		if d.JWTConnection.IssuerID == "" {
			return nil, fmt.Errorf("invalid authentication credentials, empty issuer_id in response (%s)", rawCreds)
		}
		if d.JWTConnection.KeyID == "" {
			return nil, fmt.Errorf("invalid authentication credentials, empty key_id in response (%s)", rawCreds)
		}
		if d.JWTConnection.PrivateKey == "" {
			return nil, fmt.Errorf("invalid authentication credentials, empty private_key in response (%s)", rawCreds)
		}
	}

	return &AppleDeveloperConnection{
		SessionConnection: d.SessionConnection,
		JWTConnection:     d.JWTConnection,
		TestDevices:       d.TestDevices,
	}, nil
}

func (c *BitriseClient) download() ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.buildURL, appleDeveloperConnectionPath)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for URL (%s): %s", url, err)
	}
	req.Header.Add("BUILD_API_TOKEN", c.buildAPIToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// On error, any Response can be ignored
		return nil, fmt.Errorf("failed to perform request: %s", err)
	}

	// The client must close the response body when finished with it
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Warnf("Failed to close response body: %s", cerr)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NetworkError{Status: resp.StatusCode, Body: string(body)}
	}

	return body, nil
}

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

// SessionConnection represents a Bitrise.io session-based Apple Developer connection.
type SessionConnection struct {
	AppleID              string              `json:"apple_id"`
	Password             string              `json:"password"`
	ConnectionExpiryDate string              `json:"connection_expiry_date"`
	SessionCookies       map[string][]cookie `json:"session_cookies"`
}

// JWTConnection ...
type JWTConnection struct {
	KeyID      string `json:"key_id"`
	IssuerID   string `json:"issuer_id"`
	PrivateKey string `json:"private_key"`
}

// TestDevice ...
type TestDevice struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	DeviceID   string `json:"device_identifier"`
	Title      string `json:"title"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	DeviceType string `json:"device_type"`
}

// AppleDeveloperConnection represents a Bitrise.io Apple Developer connection.
// https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/
type AppleDeveloperConnection struct {
	SessionConnection *SessionConnection
	JWTConnection     *JWTConnection
	TestDevices       []TestDevice `json:"test_devices"`
}

// PrivateKeyWithHeader adds header and footer if needed
func (cred *JWTConnection) PrivateKeyWithHeader() string {
	if strings.HasPrefix(cred.PrivateKey, "-----BEGIN PRIVATE KEY----") {
		return cred.PrivateKey
	}

	return fmt.Sprint(
		"-----BEGIN PRIVATE KEY-----\n",
		cred.PrivateKey,
		"\n-----END PRIVATE KEY-----",
	)
}

// Expiry returns the expiration of the Bitrise session-based Apple Developer connection.
func (c *SessionConnection) Expiry() *time.Time {
	t, err := time.Parse(time.RFC3339, c.ConnectionExpiryDate)
	if err != nil {
		log.Warnf("Could not parse session-based connection expiry date: %s", err)
		return nil
	}
	return &t
}

// Expired returns whether the Bitrise session-based Apple Developer connection is expired.
func (c *SessionConnection) Expired() bool {
	expiry := c.Expiry()
	if expiry == nil {
		return false
	}
	return expiry.Before(time.Now())
}

// FastlaneLoginSession returns the Apple ID login session in a ruby/object:HTTP::Cookie format.
// The session can be used as a value for FASTLANE_SESSION environment variable: https://docs.fastlane.tools/best-practices/continuous-integration/#two-step-or-two-factor-auth.
func (c *SessionConnection) FastlaneLoginSession() (string, error) {
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
