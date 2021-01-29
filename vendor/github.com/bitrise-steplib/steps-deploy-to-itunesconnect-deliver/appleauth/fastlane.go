package appleauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
)

// fastlaneAPIKey is used to serialize App Store Connect API Key into JSON for fastlane
// see: https://docs.fastlane.tools/app-store-connect-api/#using-fastlane-api-key-json-file
type fastlaneAPIKey struct {
	KeyID      string `json:"key_id"`
	IssuerID   string `json:"issuer_id"`
	PrivateKey string `json:"key"`
}

// FastlaneParams are Fastlane command arguments and environment variables
type FastlaneParams struct {
	Envs, Args []string
}

// AppendFastlaneCredentials adds auth credentials to Fastlane envs and args
func AppendFastlaneCredentials(p FastlaneParams, authConfig Credentials) error {
	if authConfig.AppleID != nil {
		// Set as environment variables
		if authConfig.AppleID.Password != "" {
			p.Envs = append(p.Envs, "DELIVER_PASSWORD="+authConfig.AppleID.Password)
		}

		if authConfig.AppleID.Session != "" {
			p.Envs = append(p.Envs, "FASTLANE_SESSION="+authConfig.AppleID.Session)
		}

		if authConfig.AppleID.AppSpecificPassword != "" {
			p.Envs = append(p.Envs, "FASTLANE_APPLE_APPLICATION_SPECIFIC_PASSWORD="+authConfig.AppleID.AppSpecificPassword)
		}

		// Add as an argument
		if authConfig.AppleID.Username != "" {
			usernameKey := "--username"
			if !sliceutil.IsStringInSlice(usernameKey, p.Args) {
				p.Args = append(p.Args, usernameKey, authConfig.AppleID.Username)
			}
		}
		if authConfig.AppleID.TeamName != "" {
			teamNameKey := "--team_name"
			if !sliceutil.IsStringInSlice(teamNameKey, p.Args) {
				p.Args = append(p.Args, teamNameKey, authConfig.AppleID.TeamName)
			}
		}
		if authConfig.AppleID.TeamID != "" {
			teamIDKey := "--team_id"
			if !sliceutil.IsStringInSlice(teamIDKey, p.Args) {
				p.Args = append(p.Args, teamIDKey, authConfig.AppleID.TeamID)
			}
		}
	}

	if authConfig.APIKey != nil {
		fastlaneAuthFile, err := writeFastlaneAPIKeyToFile(fastlaneAPIKey{
			IssuerID:   authConfig.APIKey.IssuerID,
			KeyID:      authConfig.APIKey.KeyID,
			PrivateKey: authConfig.APIKey.PrivateKey,
		})
		if err != nil {
			return fmt.Errorf("failed to write Fastane API Key configuration to file: %v", err)
		}

		apiKeyPathKey := "--api_key_path"
		precheckIAPKey := "--precheck_include_in_app_purchases"
		if !sliceutil.IsStringInSlice(apiKeyPathKey, p.Args) && !sliceutil.IsStringInSlice(precheckIAPKey, p.Args) {
			p.Args = append(p.Args, apiKeyPathKey, fastlaneAuthFile)
			// deliver: "Precheck cannot check In-app purchases with the App Store Connect API Key (yet). Exclude In-app purchases from precheck"
			p.Args = append(p.Args, precheckIAPKey, "false")
		}
	}

	return nil
}

// writeFastlaneAPIKeyToFile writes a Fastlane-specific JSON file to disk, containing Apple Service authentication details
func writeFastlaneAPIKeyToFile(authData fastlaneAPIKey) (string, error) {
	json, err := json.Marshal(authData)
	if err != nil {
		return "", err
	}

	tmpDir, err := pathutil.NormalizedOSTempDirPath("apiKey")
	if err != nil {
		return "", err
	}
	tmpPath := filepath.Join(tmpDir, "api_key.json")

	if err := ioutil.WriteFile(tmpPath, json, os.ModePerm); err != nil {
		return "", err
	}

	return tmpPath, nil
}
