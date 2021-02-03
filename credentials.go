package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/steps-deploy-to-itunesconnect-deliver/appleauth"
)

// fastlaneAPIKey is used to serialize App Store Connect API Key into JSON for fastlane
// see: https://docs.fastlane.tools/app-store-connect-api/#using-fastlane-api-key-json-file
type fastlaneAPIKey struct {
	KeyID      string `json:"key_id"`
	IssuerID   string `json:"issuer_id"`
	PrivateKey string `json:"key"`
}

// FastlaneAuthParams converts Apple credentials to Fastlane env vars and arguments
func FastlaneAuthParams(authConfig appleauth.Credentials) (map[string]string, error) {
	envs := make(map[string]string)
	if authConfig.AppleID != nil {
		// Set as environment variables
		if authConfig.AppleID.Username != "" {
			envs["DELIVER_USERNAME"] = authConfig.AppleID.Username
		}
		// FASTLANE_USER

		if authConfig.AppleID.Password != "" {
			envs["DELIVER_PASSWORD"] = authConfig.AppleID.Password
		}
		// FASTLANE_PASSWORD

		if authConfig.AppleID.Session != "" {
			envs["FASTLANE_SESSION"] = authConfig.AppleID.Session
		}

		if authConfig.AppleID.AppSpecificPassword != "" {
			envs["FASTLANE_APPLE_APPLICATION_SPECIFIC_PASSWORD"] = authConfig.AppleID.AppSpecificPassword
		}
	}

	if authConfig.APIKey != nil {
		fastlaneAPIKeyParams, err := json.Marshal(fastlaneAPIKey{
			IssuerID:   authConfig.APIKey.IssuerID,
			KeyID:      authConfig.APIKey.KeyID,
			PrivateKey: authConfig.APIKey.PrivateKey,
		})
		if err != nil {
			return envs, fmt.Errorf("failed to marshal Fastane API Key configuration: %v", err)
		}

		tmpDir, err := pathutil.NormalizedOSTempDirPath("apiKey")
		if err != nil {
			return envs, err
		}
		fastlaneAuthFile := filepath.Join(tmpDir, "api_key.json")
		if err := ioutil.WriteFile(fastlaneAuthFile, fastlaneAPIKeyParams, os.ModePerm); err != nil {
			return envs, err
		}

		envs["DELIVER_API_KEY_PATH"] = fastlaneAuthFile
		// deliver: "Precheck cannot check In-app purchases with the App Store Connect API Key (yet). Exclude In-app purchases from precheck"
		envs["PRECHECK_INCLUDE_IN_APP_PURCHASES"] = "false"
	}

	return envs, nil
}
