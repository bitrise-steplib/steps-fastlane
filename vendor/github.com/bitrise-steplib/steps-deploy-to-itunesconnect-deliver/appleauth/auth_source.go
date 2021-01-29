package appleauth

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-deploy-to-itunesconnect-deliver/devportalservice"
)

// Source returns a specific kind (Apple ID/API Key) Apple authentication data from a specific source (Bitrise Service, manual input)
type Source interface {
	Fetch(connection *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error)
	Description() string
	RequiresConnection() bool
}

// ConnectionAPIKeySource provides API Key from Bitrise Service
type ConnectionAPIKeySource struct{}

// InputAPIKeySource provides API Key from manual input
type InputAPIKeySource struct{}

// ConnectionAppleIDSource provides Apple ID from Bitrise Service
type ConnectionAppleIDSource struct{}

// InputAppleIDSource provides Apple ID from manual input
type InputAppleIDSource struct{}

// Description ...
func (*ConnectionAPIKeySource) Description() string {
	return "Connected Apple Developer Portal Account for App Store Connect API found"
}

// RequiresConnection ...
func (*ConnectionAPIKeySource) RequiresConnection() bool {
	return true
}

// Fetch ...
func (*ConnectionAPIKeySource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if conn == nil || conn.JWTConnection == nil { // Not configured
		return nil, nil
	}

	return &Credentials{
		APIKey: conn.JWTConnection,
	}, nil
}

//

// Description ...
func (*InputAPIKeySource) Description() string {
	return "Authenticating using Step inputs (App Store Connect API)"
}

// RequiresConnection ...
func (*InputAPIKeySource) RequiresConnection() bool {
	return false
}

// Fetch ...
func (*InputAPIKeySource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if inputs.APIKeyPath == "" { // Not configured
		return nil, nil
	}

	privateKey, keyID, err := fetchPrivateKey(inputs.APIKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not fetch private key (%s) specified as input: %v", inputs.APIKeyPath, err)
	}
	if len(privateKey) == 0 {
		return nil, fmt.Errorf("private key (%s) is empty", inputs.APIKeyPath)
	}

	return &Credentials{
		APIKey: &devportalservice.JWTConnection{
			IssuerID:   inputs.APIIssuer,
			KeyID:      keyID,
			PrivateKey: string(privateKey),
		},
	}, nil
}

//

// Description ...
func (*ConnectionAppleIDSource) Description() string {
	return "Connected session-based Apple Developer Portal Account found. It is Reccomended to use an API Key (App Store Connect API) instead."
}

// RequiresConnection ...
func (*ConnectionAppleIDSource) RequiresConnection() bool {
	return true
}

// Fetch ...
func (*ConnectionAppleIDSource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if conn == nil || conn.SessionConnection == nil { // No Apple ID configured
		return nil, nil
	}

	sessionConn := conn.SessionConnection
	if expiry := sessionConn.Expiry(); expiry != nil && sessionConn.Expired() {
		log.Warnf("TFA session expired on %s.", expiry.String())
		return nil, nil
	}
	session, err := sessionConn.FastlaneLoginSession()
	if err != nil {
		handleSessionDataError(err)
		return nil, nil
	}

	return &Credentials{
		AppleID: &AppleID{
			Username:            conn.SessionConnection.AppleID,
			Password:            conn.SessionConnection.Password,
			Session:             session,
			AppSpecificPassword: inputs.AppSpecificPassword,
			TeamID:              inputs.TeamID,
			TeamName:            inputs.TeamName,
		},
	}, nil
}

//

// Description ...
func (*InputAppleIDSource) Description() string {
	return "Authenticating using Step inputs (session-based). This method does not support 2FA. It is Reccomended to use an API Key (App Store Connect API) instead."
}

// RequiresConnection ...
func (*InputAppleIDSource) RequiresConnection() bool {
	return false
}

// Fetch ...
func (*InputAppleIDSource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if inputs.Username == "" { // Not configured
		return nil, nil
	}

	return &Credentials{
		AppleID: &AppleID{
			Username:            inputs.Username,
			Password:            inputs.Password,
			AppSpecificPassword: inputs.AppSpecificPassword,
			TeamID:              inputs.TeamID,
			TeamName:            inputs.TeamName,
		},
	}, nil
}
