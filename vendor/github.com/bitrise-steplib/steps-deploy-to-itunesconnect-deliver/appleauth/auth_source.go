package appleauth

import (
	"fmt"

	"github.com/bitrise-steplib/steps-deploy-to-itunesconnect-deliver/devportalservice"
)

// Source returns a specific kind (Apple ID/API Key) Apple authentication data from a specific source (Bitrise Apple Developer Connection, Step inputs)
type Source interface {
	Fetch(connection *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error)
	Description() string
}

// ConnectionAPIKeySource provides API Key from Bitrise Apple Developer Connection
type ConnectionAPIKeySource struct{}

// InputAPIKeySource provides API Key from Step inputs
type InputAPIKeySource struct{}

// ConnectionAppleIDSource provides Apple ID from Bitrise Apple Developer Connection
type ConnectionAppleIDSource struct{}

// InputAppleIDSource provides Apple ID from Step inputs
type InputAppleIDSource struct{}

// ConnectionAppleIDFastlaneSource provides Apple ID from Bitrise Apple Developer Connection, includes Fastlane specific session
type ConnectionAppleIDFastlaneSource struct{}

// InputAppleIDFastlaneSource provides Apple ID from Step inputs, includes Fastlane specific session
type InputAppleIDFastlaneSource struct{}

// Description ...
func (*ConnectionAPIKeySource) Description() string {
	return "Bitrise Apple Developer Connection with API key found"
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
	return "Inputs with API key authentication found"
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
	return "Bitrise Apple Developer Connection with Apple ID found."
}

// Fetch ...
func (*ConnectionAppleIDSource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if conn == nil || conn.SessionConnection == nil { // No Apple ID configured
		return nil, nil
	}

	return &Credentials{
		AppleID: &AppleID{
			Username:            conn.SessionConnection.AppleID,
			Password:            conn.SessionConnection.Password,
			Session:             "",
			AppSpecificPassword: inputs.AppSpecificPassword,
		},
	}, nil
}

//

// Description ...
func (*InputAppleIDSource) Description() string {
	return "Inputs with Apple ID authentication found."
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
		},
	}, nil
}

//

// Description ...
func (*ConnectionAppleIDFastlaneSource) Description() string {
	return "Bitrise Apple Developer Connection with Apple ID found."
}

// Fetch ...
func (*ConnectionAppleIDFastlaneSource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if conn == nil || conn.SessionConnection == nil { // No Apple ID configured
		return nil, nil
	}

	sessionConn := conn.SessionConnection
	if expiry := sessionConn.Expiry(); expiry != nil && sessionConn.Expired() {
		return nil, fmt.Errorf("2FA session saved in Bitrise Developer Connection is expired, was valid until %s", expiry.String())
	}
	session, err := sessionConn.FastlaneLoginSession()
	if err != nil {
		return nil, fmt.Errorf("could not prepare Fastlane session cookie object: %v", err)
	}

	return &Credentials{
		AppleID: &AppleID{
			Username:            conn.SessionConnection.AppleID,
			Password:            conn.SessionConnection.Password,
			Session:             session,
			AppSpecificPassword: inputs.AppSpecificPassword,
		},
	}, nil
}

//

// Description ...
func (*InputAppleIDFastlaneSource) Description() string {
	return "Inputs with Apple ID authentication found. This method does not support TFA enabled Apple IDs."
}

// Fetch ...
func (*InputAppleIDFastlaneSource) Fetch(conn *devportalservice.AppleDeveloperConnection, inputs Inputs) (*Credentials, error) {
	if inputs.Username == "" { // Not configured
		return nil, nil
	}

	return &Credentials{
		AppleID: &AppleID{
			Username:            inputs.Username,
			Password:            inputs.Password,
			AppSpecificPassword: inputs.AppSpecificPassword,
		},
	}, nil
}
