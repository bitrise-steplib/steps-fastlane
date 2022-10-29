package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-xcode/appleauth"
	"github.com/bitrise-io/go-xcode/devportalservice"
	"github.com/kballard/go-shellquote"
)

// Config contains inputs parsed from environment variables
type Config struct {
	WorkDir string `env:"work_dir,dir"`
	Lane    string `env:"lane,required"`

	BitriseConnection   string          `env:"connection,opt[automatic,api_key,apple_id,off]"`
	AppleID             string          `env:"apple_id"`
	Password            stepconf.Secret `env:"password"`
	AppSpecificPassword stepconf.Secret `env:"app_password"`
	APIKeyPath          stepconf.Secret `env:"api_key_path"`
	APIIssuer           string          `env:"api_issuer"`

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
func (s FastlaneRunner) ProcessConfig() (Config, error) {
	var config Config
	if err := s.inputParser.Parse(&config); err != nil {
		return config, err
	}

	stepconf.Print(config)
	s.logger.EnableDebugLog(config.VerboseLog)
	fmt.Println()

	authInputs, err := s.validateAuthInputs(config)
	if err != nil {
		return Config{}, fmt.Errorf("Issue with authentication related inputs: %v", err)
	}

	authSources, err := s.parseAuthSources(config.BitriseConnection)
	if err != nil {
		return Config{}, fmt.Errorf("Invalid Input: %v", err)
	}

	s.validateGemHome(config)

	workDir, err := s.getWorkdir(config)
	if err != nil {
		return Config{}, err
	}
	config.WorkDir = workDir

	s.checkForRbenv(workDir)

	// Select and fetch Apple authenication source
	authConfig, err := s.selectAppleAuthSource(config, authSources, authInputs)
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

func (s FastlaneRunner) validateAuthInputs(config Config) (appleauth.Inputs, error) {
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

func (s FastlaneRunner) parseAuthSources(bitriseConnection string) ([]appleauth.Source, error) {
	switch bitriseConnection {
	case "automatic":
		return []appleauth.Source{
			&appleauth.ConnectionAPIKeySource{},
			&appleauth.ConnectionAppleIDFastlaneSource{},
			&appleauth.InputAPIKeySource{},
			&appleauth.InputAppleIDFastlaneSource{},
		}, nil
	case "api_key":
		return []appleauth.Source{&appleauth.ConnectionAPIKeySource{}}, nil
	case "apple_id":
		return []appleauth.Source{&appleauth.ConnectionAppleIDFastlaneSource{}}, nil
	case "off":
		return []appleauth.Source{
			&appleauth.InputAPIKeySource{},
			&appleauth.InputAppleIDFastlaneSource{},
		}, nil
	default:
		return nil, fmt.Errorf("invalid connection input: %s", bitriseConnection)
	}
}

func (s FastlaneRunner) validateGemHome(config Config) {
	if strings.TrimSpace(config.GemHome) == "" {
		return
	}
	s.logger.Warnf("Custom value (%s) is set for GEM_HOME environment variable. This can lead to errors as gem lookup path may not contain GEM_HOME.")
}

func (s FastlaneRunner) getWorkdir(config Config) (string, error) {
	s.logger.Infof("Expand WorkDir")

	workDir := config.WorkDir
	if workDir == "" {
		s.logger.Printf("WorkDir not set, using CurrentWorkingDirectory...")
		currentDir, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
		if err != nil {
			return "", fmt.Errorf("Failed to get current dir, error: %s", err)
		}
		workDir = currentDir
	} else {
		absWorkDir, err := pathutil.AbsPath(workDir)
		if err != nil {
			return "", fmt.Errorf("Failed to expand path (%s), error: %s", workDir, err)
		}
		workDir = absWorkDir
	}

	s.logger.Donef("Expanded WorkDir: %s", workDir)
	return workDir, nil
}

func (s FastlaneRunner) checkForRbenv(workDir string) {
	if _, err := s.cmdLocator.LookPath("rbenv"); err != nil {
		cmd := s.rbyFactory.Create("rbenv", []string{"versions"}, &command.Opts{
			Stderr: os.Stderr,
			Stdout: os.Stdout,
			Dir:    workDir,
		})

		fmt.Println()
		s.logger.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			s.logger.Warnf("%s", err)
		}
	}
}

func (s FastlaneRunner) selectAppleAuthSource(config Config, authSources []appleauth.Source, authInputs appleauth.Inputs) (appleauth.Credentials, error) {
	var devportalConnectionProvider *devportalservice.BitriseClient
	if config.BuildURL != "" && config.BuildAPIToken != "" {
		devportalConnectionProvider = devportalservice.NewBitriseClient(retry.NewHTTPClient().StandardClient(), config.BuildURL, string(config.BuildAPIToken))
	} else {
		fmt.Println()
		s.logger.Warnf("Connected Apple Developer Portal Account not found. Step is not running on bitrise.io: BITRISE_BUILD_URL and BITRISE_BUILD_API_TOKEN envs are not set")
	}
	var conn *devportalservice.AppleDeveloperConnection
	if config.BitriseConnection != "off" && devportalConnectionProvider != nil {
		var err error
		conn, err = devportalConnectionProvider.GetAppleDeveloperConnection()
		if err != nil {
			s.handleSessionDataError(err)
		}
	}

	authConfig, err := appleauth.Select(conn, authSources, authInputs)
	if err != nil {
		if _, ok := err.(*appleauth.MissingAuthConfigError); !ok {
			return appleauth.Credentials{}, fmt.Errorf("Could not configure Apple Service authentication: %v", err)
		}
		fmt.Println()
		s.logger.Warnf("No authentication data found matching the selected Apple Service authentication method (%s).", config.BitriseConnection)
		if conn != nil && (conn.APIKeyConnection == nil && conn.AppleIDConnection == nil) {
			fmt.Println()
			s.logger.Warnf("%s", notConnected)
		}
	}
	return authConfig, nil
}

const notConnected = `Connected Apple Developer Portal Account not found.
Most likely because there is no Apple Developer Portal Account connected to the build.
Read more: https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/`

func (s FastlaneRunner) handleSessionDataError(err error) {
	if err == nil {
		return
	}

	if networkErr, ok := err.(devportalservice.NetworkError); ok && networkErr.Status == http.StatusUnauthorized {
		fmt.Println()
		s.logger.Warnf("%s", "Unauthorized to query Connected Apple Developer Portal Account. This happens by design, with a public app's PR build, to protect secrets.")

		return
	}

	fmt.Println()
	s.logger.Errorf("Failed to activate Bitrise Apple Developer Portal connection: %s", err)
	s.logger.Warnf("Read more: https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/")
}