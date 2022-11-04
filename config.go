package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/kballard/go-shellquote"
)

// Config contains inputs parsed from environment variables
type Config struct {
	WorkDir string `env:"work_dir,dir"`
	Lane    string `env:"lane,required"`

	BitriseConnection   bitriseConnection `env:"connection,opt[automatic,api_key,apple_id,off]"`
	AppleID             string            `env:"apple_id"`
	Password            stepconf.Secret   `env:"password"`
	AppSpecificPassword stepconf.Secret   `env:"app_password"`
	APIKeyPath          stepconf.Secret   `env:"api_key_path"`
	APIIssuer           string            `env:"api_issuer"`

	UpdateFastlane bool `env:"update_fastlane,opt[true,false]"`
	VerboseLog     bool `env:"verbose_log,opt[yes,no]"`
	EnableCache    bool `env:"enable_cache,opt[yes,no]"`

	GemHome string `env:"GEM_HOME"`

	// Used to get Bitrise Apple Developer Portal Connection
	BuildURL        string          `env:"BITRISE_BUILD_URL"`
	BuildAPIToken   stepconf.Secret `env:"BITRISE_BUILD_API_TOKEN"`
	AuthInputs      appleauth.Inputs
	AuthCredentials appleauth.Credentials
	LaneOptions     []string
}

// ProcessConfig ...
func (f FastlaneRunner) ProcessConfig() (Config, error) {
	var config Config
	if err := f.inputParser.Parse(&config); err != nil {
		return config, err
	}

	stepconf.Print(config)
	f.logger.EnableDebugLog(config.VerboseLog)
	fmt.Println()

	authInputs, err := f.validateAuthInputs(config)
	if err != nil {
		return Config{}, fmt.Errorf("Issue with authentication related inputs: %v", err)
	}

	authSources, err := f.parseAuthSources(config.BitriseConnection)
	if err != nil {
		return Config{}, fmt.Errorf("Invalid Input: %v", err)
	}

	f.validateGemHome(config)

	workDir, err := f.getWorkdir(config)
	if err != nil {
		return Config{}, err
	}
	config.WorkDir = workDir

	f.checkForRbenv(workDir)

	// Select and fetch Apple authenication source
	authConfig, err := f.selectAppleAuthSource(config, authSources, authInputs)
	if err != nil {
		return Config{}, err
	}
	config.AuthCredentials = authConfig

	// Split lane option
	laneOptions, err := shellquote.Split(config.Lane)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to parse lane (%s), error: %s", config.Lane, err)
	}
	config.LaneOptions = laneOptions

	return config, nil
}

func (f FastlaneRunner) validateAuthInputs(config Config) (appleauth.Inputs, error) {
	authInputs := appleauth.Inputs{
		Username:            config.AppleID,
		Password:            string(config.Password),
		AppSpecificPassword: string(config.AppSpecificPassword),
		APIIssuer:           config.APIIssuer,
		APIKeyPath:          string(config.APIKeyPath),
	}
	if err := authInputs.Validate(); err != nil {
		return appleauth.Inputs{}, err
	}
	return authInputs, nil
}

type bitriseConnection string

const (
	automatic = "automatic"
	apiKey    = "api_key"
	appleId   = "apple_id"
	off       = "off"
)

func (f FastlaneRunner) parseAuthSources(connection bitriseConnection) ([]appleauth.Source, error) {
	switch connection {
	case automatic:
		return []appleauth.Source{
			&appleauth.ConnectionAPIKeySource{},
			&appleauth.ConnectionAppleIDFastlaneSource{},
			&appleauth.InputAPIKeySource{},
			&appleauth.InputAppleIDFastlaneSource{},
		}, nil
	case apiKey:
		return []appleauth.Source{&appleauth.ConnectionAPIKeySource{}}, nil
	case appleId:
		return []appleauth.Source{&appleauth.ConnectionAppleIDFastlaneSource{}}, nil
	case off:
		return []appleauth.Source{
			&appleauth.InputAPIKeySource{},
			&appleauth.InputAppleIDFastlaneSource{},
		}, nil
	default:
		return nil, fmt.Errorf("invalid connection input: %s", connection)
	}
}

func (f FastlaneRunner) validateGemHome(config Config) {
	if strings.TrimSpace(config.GemHome) == "" {
		return
	}
	f.logger.Warnf("GEM_HOME environment variable is set to:\n%s\nThis can lead to errors as gem lookup path may not contain GEM_HOME.", config.GemHome)
}

func (f FastlaneRunner) getWorkdir(config Config) (string, error) {
	f.logger.Infof("Expand WorkDir")

	workDir := config.WorkDir
	if workDir == "" {
		f.logger.Printf("WorkDir not set, using CurrentWorkingDirectory...")
		currentDir, err := f.pathModifier.AbsPath(".")
		if err != nil {
			return "", fmt.Errorf("Failed to get current dir, error: %s", err)
		}
		workDir = currentDir
	} else {
		absWorkDir, err := f.pathModifier.AbsPath(workDir)
		if err != nil {
			return "", fmt.Errorf("Failed to expand path (%s), error: %s", workDir, err)
		}
		workDir = absWorkDir
	}

	f.logger.Donef("Expanded WorkDir: %s", workDir)
	return workDir, nil
}

func (f FastlaneRunner) checkForRbenv(workDir string) {
	if _, err := f.cmdLocator.LookPath("rbenv"); err != nil {
		cmd := f.rbyFactory.Create("rbenv", []string{"versions"}, &command.Opts{
			Stderr: os.Stderr,
			Stdout: os.Stdout,
			Dir:    workDir,
		})

		fmt.Println()
		f.logger.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			f.logger.Warnf("%s", err)
		}
	}
}

func (f FastlaneRunner) selectAppleAuthSource(config Config, authSources []appleauth.Source, authInputs appleauth.Inputs) (appleauth.Credentials, error) {
	var devportalConnectionProvider *devportalservice.BitriseClient
	if config.BuildURL != "" && config.BuildAPIToken != "" {
		devportalConnectionProvider = devportalservice.NewBitriseClient(retryhttp.NewClient(f.logger).StandardClient(), config.BuildURL, string(config.BuildAPIToken))
	} else {
		fmt.Println()
		f.logger.Warnf("Connected Apple Developer Portal Account not found. Step is not running on bitrise.io: BITRISE_BUILD_URL and BITRISE_BUILD_API_TOKEN envs are not set")
	}
	var conn *devportalservice.AppleDeveloperConnection
	if config.BitriseConnection != "off" && devportalConnectionProvider != nil {
		var err error
		conn, err = devportalConnectionProvider.GetAppleDeveloperConnection()
		if err != nil {
			f.handleSessionDataError(err)
		}
	}

	authConfig, err := appleauth.Select(conn, authSources, authInputs)
	if err != nil {
		if _, ok := err.(*appleauth.MissingAuthConfigError); !ok {
			return appleauth.Credentials{}, fmt.Errorf("Could not configure Apple Service authentication: %v", err)
		}
		fmt.Println()
		f.logger.Warnf("No authentication data found matching the selected Apple Service authentication method (%s).", config.BitriseConnection)
		if conn != nil && (conn.APIKeyConnection == nil && conn.AppleIDConnection == nil) {
			fmt.Println()
			f.logger.Warnf("%s", notConnected)
		}
	}
	return authConfig, nil
}

const notConnected = `Connected Apple Developer Portal Account not found.
Most likely because there is no Apple Developer Portal Account connected to the build.
Read more: https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/`

func (f FastlaneRunner) handleSessionDataError(err error) {
	if err == nil {
		return
	}

	if networkErr, ok := err.(devportalservice.NetworkError); ok && networkErr.Status == http.StatusUnauthorized {
		fmt.Println()
		f.logger.Warnf("%s", "Unauthorized to query Connected Apple Developer Portal Account. This happens by design, with a public app's PR build, to protect secrets.")

		return
	}

	fmt.Println()
	f.logger.Errorf("Failed to activate Bitrise Apple Developer Portal connection: %s", err)
	f.logger.Warnf("Read more: https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/")
}
