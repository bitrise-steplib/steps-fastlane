package devportalservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/bitrise-io/go-utils/log"
)

// NetworkError ...
type NetworkError struct {
	Status int
	Body   string
}

// portalData ...
type portalData struct {
	AppleID              string              `json:"apple_id"`
	Password             string              `json:"password"`
	ConnectionExpiryDate string              `json:"connection_expiry_date"`
	SessionCookies       map[string][]cookie `json:"session_cookies"`
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

// SessionData will fetch the session from Bitrise for the connected Apple developer account
// If the BITRISE_PORTAL_DATA_JSON is provided (for debug purposes) it will use that instead.
func SessionData() (string, error) {
	p, err := getDeveloperPortalData(os.Getenv("BITRISE_BUILD_URL"), os.Getenv("BITRISE_BUILD_API_TOKEN"))
	if err != nil {
		return "", err
	}

	cookies, err := convertDesCookie(p.SessionCookies["https://idmsa.apple.com"])
	if err != nil {
		return "", err
	}
	return strings.Join(cookies, ""), nil
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("response %d %s", e.Status, e.Body)
}

func getDeveloperPortalData(buildURL, buildAPIToken string) (portalData, error) {
	var p portalData

	j, exists := os.LookupEnv("BITRISE_PORTAL_DATA_JSON")
	if exists && j != "" {
		return p, json.Unmarshal([]byte(j), &p)
	}

	if buildURL == "" {
		return portalData{}, fmt.Errorf("BITRISE_BUILD_URL env is not exported")
	}

	if buildAPIToken == "" {
		return portalData{}, fmt.Errorf("BITRISE_BUILD_API_TOKEN env is not exported")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/apple_developer_portal_data.json", buildURL), nil)
	if err != nil {
		return portalData{}, err
	}

	req.Header.Add("BUILD_API_TOKEN", buildAPIToken)

	if _, err := performRequest(req, &p); err != nil {
		return portalData{}, fmt.Errorf("Falied to fetch portal data from Bitrise, error: %s", err)
	}
	return p, nil
}

func convertDesCookie(cookies []cookie) ([]string, error) {
	var convertedCookies []string
	var errs []string
	for _, c := range cookies {
		if convertedCookies == nil {
			convertedCookies = append(convertedCookies, "---"+"\n")
		}

		if c.ForDomain == nil {
			b := true
			c.ForDomain = &b
		}

		tmpl, err := template.New("").Parse(`- !ruby/object:HTTP::Cookie
  name: {{.Name}}
  value: {{.Value}}
  domain: {{.Domain}}
  for_domain: {{.ForDomain}}
  path: "{{.Path}}"
`)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Failed to create golang template for the cookie: %v", c))
			continue
		}

		var b bytes.Buffer
		err = tmpl.Execute(&b, c)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Failed to parse cookie: %v", c))
			continue
		}

		convertedCookies = append(convertedCookies, b.String()+"\n")
	}

	return convertedCookies, errors.New(strings.Join(errs, "\n"))
}

func performRequest(req *http.Request, requestResponse interface{}) ([]byte, error) {
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		// On error, any Response can be ignored
		return nil, fmt.Errorf("failed to perform request, error: %s", err)
	}

	// The client must close the response body when finished with it
	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			log.Warnf("Failed to close response body, error: %s", cerr)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body, error: %s", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, NetworkError{Status: response.StatusCode, Body: string(body)}
	}

	// Parse JSON body
	if requestResponse != nil {
		if err := json.Unmarshal([]byte(body), &requestResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response (%s), error: %s", body, err)
		}
	}
	return body, nil
}

func main() {
}
