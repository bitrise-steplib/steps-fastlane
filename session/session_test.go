package session

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestGetDeveloperPortalDataFromEnv(t *testing.T) {
	// If the BITRISE_PORTAL_DATA_JSON env is already set (e.g for local testing the step),
	// then we don't want to modify it during the test
	var envAlreadyPresented bool
	if os.Getenv("BITRISE_PORTAL_DATA_JSON") == "" {
		if err := os.Setenv("BITRISE_PORTAL_DATA_JSON", dummyPortalDataJSON); err != nil {
			t.Errorf("Failed to set BITRISE_PORTAL_DATA_JSON env to test getDeveloperPortalData() error = %v", err)
		}
	} else {
		envAlreadyPresented = true
	}

	got, err := getDeveloperPortalData("", "")
	if err != nil {
		t.Errorf("getDeveloperPortalData() error = %v", err)
		return
	}

	if !envAlreadyPresented {
		if !reflect.DeepEqual(got, dummyPortalData) {
			t.Errorf("getDeveloperPortalData() = %v, want %v", sPretty(got), sPretty(dummyPortalData))
		}

		// Reset the dummy BITRISE_PORTAL_DATA_JSON
		if err := os.Setenv("BITRISE_PORTAL_DATA_JSON", ""); err != nil {
			t.Errorf("Failed to reset BITRISE_PORTAL_DATA_JSON env after testing getDeveloperPortalData() error = %v", err)
		}
	}
}

// TestGetDeveloperPortalData
func TestGetDeveloperPortalData(t *testing.T) {
	buildURL, builAPIToken := os.Getenv("BITRISE_BUILD_URL"), os.Getenv("BITRISE_BUILD_API_TOKEN")
	if buildURL == "" {
		t.Skip("Failed to run TestGetDeveloperPortalData() test because the BITRISE_BUILD_URL env is not exported")
	}
	if builAPIToken == "" {
		t.Skip("Failed to run TestGetDeveloperPortalData() test because the BITRISE_BUILD_API_TOKEN env is not exported")
	}

	got, err := getDeveloperPortalData(buildURL, builAPIToken)
	if err != nil {
		t.Errorf("getDeveloperPortalData() error = %v", err)
	}

	if reflect.DeepEqual(got, PortalData{}) {
		t.Errorf("getDeveloperPortalData() = nil")
	}

	t.Logf(sPretty(got))
}

const dummyPortalDataJSON = `{
    "apple_id": "example@example.io",
    "password": "highSecurityPassword",
    "connection_expiry_date": "2019-04-06T12:04:59.000Z",
    "session_cookies": {
        "https://idmsa.apple.com": [
            {
                "name": "DES58b0eba556d80ed2b98707e15ffafd344",
                "path": "/",
                "value": "HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT",
                "domain": "idmsa.apple.com",
                "secure": true,
                "expires": "2019-04-06T12:04:59Z",
                "max_age": 2592000,
                "httponly": true
            },
            {
                "name": "myacinfo",
                "path": "/",
                "value": "DAWTKNV26a0a6db3ae43acd203d0d03e8bc45000cd4bdc668e90953f22ca3b36eaab0e18634660a10cf28cc65d8ddf633c017de09477dfb18c8a3d6961f96cbbf064be616e80cee62d3d7f39a485bf826377c5b5dbbfc4a97dcdb462052db73a3a1d9b4a325d5bdd496190b3088878cecce17e4d6db9230e0575cfbe7a8754d1de0c937080ef84569b6e4a75237c2ec01cf07db060a11d92e7220707dd00a2a565ee9e06074d8efa6a1b7f83db3e1b2acdafb5fc0708443e77e6d71e168ae2a83b848122264b2da5cadfd9e451f9fe3f6eebc71904d4bc36acc528cc2a844d4f2eb527649a69523756ec9955457f704c28a3b6b9f97d6df900bd60044d5bc50408260f096954f03c53c16ac40a796dc439b859f882a50390b1c7517a9f4479fb1ce9ba2db241d6b8f2eb127c46ef96e0ccccccccc",
                "domain": "apple.com",
                "secure": true,
                "httponly": true
            }
        ]
    },
    "test_devices": [
        {
            "id": 8414,
            "user_id": 52411,
            "device_identifier": "1b78ac4bad2e8911139287ac5dd152fbe86eb2b9",
            "title": "iPhone 7",
            "created_at": "2018-08-30T09:09:36.332Z",
            "updated_at": "2018-08-30T09:09:36.332Z",
            "device_type": "ios"
        }
    ],
    "default_team_id": null
}`

var dummyPortalData = PortalData{
	AppleID:              "example@example.io",
	Password:             "highSecurityPassword",
	ConnectionExpiryDate: "2019-04-06T12:04:59.000Z",
	SessionCookies: map[string][]Cookie{
		"https://idmsa.apple.com": []Cookie{
			Cookie{
				Name:     "DES58b0eba556d80ed2b98707e15ffafd344",
				Path:     "/",
				Value:    "HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT",
				Domain:   "idmsa.apple.com",
				Secure:   true,
				Expires:  "2019-04-06T12:04:59Z",
				MaxAge:   2592000,
				Httponly: true,
			},
			Cookie{
				Name:     "myacinfo",
				Path:     "/",
				Value:    "DAWTKNV26a0a6db3ae43acd203d0d03e8bc45000cd4bdc668e90953f22ca3b36eaab0e18634660a10cf28cc65d8ddf633c017de09477dfb18c8a3d6961f96cbbf064be616e80cee62d3d7f39a485bf826377c5b5dbbfc4a97dcdb462052db73a3a1d9b4a325d5bdd496190b3088878cecce17e4d6db9230e0575cfbe7a8754d1de0c937080ef84569b6e4a75237c2ec01cf07db060a11d92e7220707dd00a2a565ee9e06074d8efa6a1b7f83db3e1b2acdafb5fc0708443e77e6d71e168ae2a83b848122264b2da5cadfd9e451f9fe3f6eebc71904d4bc36acc528cc2a844d4f2eb527649a69523756ec9955457f704c28a3b6b9f97d6df900bd60044d5bc50408260f096954f03c53c16ac40a796dc439b859f882a50390b1c7517a9f4479fb1ce9ba2db241d6b8f2eb127c46ef96e0ccccccccc",
				Domain:   "apple.com",
				Secure:   true,
				Httponly: true,
			},
		},
	},
}

func Test_convertDesCookie(t *testing.T) {
	tests := []struct {
		name    string
		cookies []Cookie
		want    []string
	}{
		{
			name: "Convert one cookie",
			cookies: []Cookie{
				Cookie{
					Name:     "DES58b0eba556d80ed2b98707e15ffafd344",
					Path:     "/",
					Value:    "HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT",
					Domain:   "idmsa.apple.com",
					Secure:   true,
					Expires:  "2019-04-06T12:04:59Z",
					MaxAge:   2592000,
					Httponly: true,
				},
			},
			want: []string{"---\n",
				"- !ruby/object:HTTP::Cookie\n" +
					"  name: DES58b0eba556d80ed2b98707e15ffafd344\n" +
					"  value: HSARMTKNSRVTWFlaFrGQTmfmFBwJuiX/aaaaaaaaa+A7FbJa4V8MmWijnJknnX06ME0KrI9V8vFg==SRVT\n" +
					"  domain: idmsa.apple.com\n" +
					"  for_domain: true\n" +
					`  path: "/"` + "\n" +
					"\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertDesCookie(tt.cookies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertDesCookie() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func sPretty(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}

	return fmt.Sprintf("%v\n", string(b))
}
