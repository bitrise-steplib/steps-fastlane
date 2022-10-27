package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-steputils/command/gems"
	"github.com/bitrise-io/go-steputils/ruby"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
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

func main() {
	os.Exit(run())
}

func run() int {
	logger := log.NewLogger()
	buildStep := createStep(logger)

	config, err := buildStep.ProcessConfig()
	if err != nil {
		buildStep.logger.Errorf(fmt.Errorf("Failed to process Step inputs: %w", err).Error())
		return 1
	}

	// Determine desired Fastlane version
	fmt.Println()
	buildStep.logger.Infof("Determine desired Fastlane version")
	gemVersions, err := parseGemfileLock(config.WorkDir)
	if err != nil {
		buildStep.logger.Errorf(err.Error())
	}

	fmt.Println()

	dependenciesOpts := EnsureDependenciesOpts{
		GemVersions: gemVersions,
		UseBundler:  gemVersions.fastlane.Found,
	}
	if err = buildStep.InstallDependencies(config, dependenciesOpts); err != nil {
		buildStep.logger.Errorf(fmt.Errorf("Failed to install Step dependencies: %w", err).Error())
		return 1
	}

	buildStep.Run(config, dependenciesOpts)

	if config.EnableCache {
		fmt.Println()
		buildStep.logger.Infof("Collecting cache")

		c := cache.New()
		for _, depFunc := range depsFuncs {
			includes, excludes, err := depFunc(config.WorkDir)
			buildStep.logger.Debugf("%s found include path:\n%s\nexclude paths:\n%s", functionName(depFunc), strings.Join(includes, "\n"), strings.Join(excludes, "\n"))
			if err != nil {
				buildStep.logger.Warnf("failed to collect dependencies: %s", err.Error())
				continue
			}

			for _, item := range includes {
				c.IncludePath(item)
			}

			for _, item := range excludes {
				c.ExcludePath(item)
			}
		}
		if err := c.Commit(); err != nil {
			buildStep.logger.Warnf("failed to commit paths to cache: %s", err)
		}
	}

	return 0
}

func createStep(logger log.Logger) FastlaneRunner {
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepository)
	cmdFactory := command.NewFactory(envRepository)
	cmdLocator := env.NewCommandLocator()
	rbyFactory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), cmdLocator)
	if err != nil {
		logger.Warnf("%s", err)
	}

	return NewFastlaneRunner(inputParser, logger, cmdLocator, cmdFactory, rbyFactory)
}

// FastlaneRunner ...
type FastlaneRunner struct {
	inputParser stepconf.InputParser
	logger      log.Logger
	cmdFactory  command.Factory
	cmdLocator  env.CommandLocator
	rbyFactory  ruby.CommandFactory
}

// Step Constructor
func NewFastlaneRunner(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	commandLocator env.CommandLocator,
	cmdFactory command.Factory,
	rbyFactory ruby.CommandFactory,
) FastlaneRunner {
	return FastlaneRunner{
		inputParser: stepInputParser,
		logger:      logger,
		cmdLocator:  commandLocator,
		cmdFactory:  cmdFactory,
		rbyFactory:  rbyFactory,
	}
}

// Process Config
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
		return appleauth.Inputs{}, fmt.Errorf("Issue with authentication related inputs: %v", err)
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

type EnsureDependenciesOpts struct {
	GemVersions gemVersions
	UseBundler  bool
}

// Install Dependencies
func (s FastlaneRunner) InstallDependencies(config Config, opts EnsureDependenciesOpts) error {
	// Install desired Fastlane version
	if opts.UseBundler {
		s.logger.Infof("Install bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bundler _1.2.3_" can return 'Command not found', installing bundler solves this
		cmds := s.rbyFactory.CreateGemInstall("bundler", opts.GemVersions.bundler.Version, false, true, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    config.WorkDir,
		})
		for _, cmd := range cmds {
			s.logger.Donef("$ %s", cmd.PrintableCommandArgs())
			fmt.Println()

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("command failed, error: %s", err)
			}
		}

		// install Gemfile.lock gems with `bundle [_version_] install ...`
		fmt.Println()
		s.logger.Infof("Install Fastlane with bundler")

		cmd := s.rbyFactory.CreateBundleInstall(opts.GemVersions.bundler.Version, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    config.WorkDir,
		})

		s.logger.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Command failed, error: %s", err)
		}
	} else if config.UpdateFastlane {
		s.logger.Infof("Update system installed Fastlane")

		cmds := s.rbyFactory.CreateGemInstall("fastlane", "", false, false, &command.Opts{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  nil,
			Env:    []string{},
			Dir:    config.WorkDir,
		})
		for _, cmd := range cmds {
			s.logger.Donef("$ %s", cmd.PrintableCommandArgs())

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("Command failed, error: %s", err)
			}
		}
	} else {
		s.logger.Infof("Using system installed Fastlane")
	}

	fmt.Println()
	s.logger.Infof("Fastlane version")

	name := "fastlane"
	args := []string{"--version"}
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  nil,
		Env:    []string{},
		Dir:    config.WorkDir,
	}
	var cmd command.Command
	if opts.UseBundler {
		cmd = s.rbyFactory.CreateBundleExec(name, args, opts.GemVersions.bundler.Version, options)
	} else {
		cmd = s.rbyFactory.Create(name, args, options)
	}

	s.logger.Donef("$ %s", cmd.PrintableCommandArgs())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Command failed, error: %s", err)
	}

	return nil
}

// Run Command
func (s FastlaneRunner) Run(config Config, opts EnsureDependenciesOpts) error {
	// Run fastlane
	fmt.Println()
	s.logger.Infof("Run Fastlane")

	var envs []string
	authEnvs, err := FastlaneAuthParams(config.AuthCredentials)
	if err != nil {
		return fmt.Errorf("Failed to set up Fastlane authentication parameters: %v", err)
	}
	var globallySetAuthEnvs []string
	for envKey, envValue := range authEnvs {
		if _, set := os.LookupEnv(envKey); set {
			globallySetAuthEnvs = append(globallySetAuthEnvs, envKey)
		}

		envs = append(envs, fmt.Sprintf("%s=%s", envKey, envValue))
	}
	if len(globallySetAuthEnvs) != 0 {
		s.logger.Warnf("Fastlane authentication-related environment varibale(s) (%s) are set, overriding.", globallySetAuthEnvs)
		s.logger.Infof("To stop overriding authentication-related environment variables, please set Bitrise Apple Developer Connection input to 'off' and leave authentication-related inputs empty.")
	}

	buildlogPth := ""
	if tempDir, err := pathutil.NormalizedOSTempDirPath("fastlane_logs"); err != nil {
		s.logger.Errorf("Failed to create temp dir for fastlane logs, error: %s", err)
	} else {
		buildlogPth = tempDir
		envs = append(envs, "FL_BUILDLOG_PATH="+buildlogPth)
	}

	name := "fastlane"
	args := config.LaneOptions
	options := &command.Opts{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    config.WorkDir,
		Env:    append(os.Environ(), envs...),
	}
	var cmd command.Command
	if opts.UseBundler {
		cmd = s.rbyFactory.CreateBundleExec(name, args, opts.GemVersions.bundler.Version, options)
	} else {
		cmd = s.rbyFactory.Create(name, args, options)
	}

	s.logger.Donef("$ %s", cmd.PrintableCommandArgs())

	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")
	if deployDir == "" {
		s.logger.Warnf("No BITRISE_DEPLOY_DIR found")
	}
	deployPth := filepath.Join(deployDir, "fastlane_env.log")

	if err := cmd.Run(); err != nil {
		fmt.Println()
		s.logger.Errorf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
		s.logger.Errorf("If you want to send an issue report to fastlane (https://github.com/fastlane/fastlane/issues/new), you can find the output of fastlane env in the following log file:")
		fmt.Println()
		s.logger.Infof(deployPth)
		fmt.Println()

		if fastlaneDebugInfo, err := s.fastlaneDebugInfo(config.WorkDir, opts.UseBundler, opts.GemVersions.bundler); err != nil {
			s.logger.Warnf("%s", err)
		} else if fastlaneDebugInfo != "" {
			if err := fileutil.WriteStringToFile(deployPth, fastlaneDebugInfo); err != nil {
				s.logger.Warnf("Failed to write fastlane env log file, error: %s", err)
			}
		}

		if err := filepath.Walk(buildlogPth, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if relLogPath, err := filepath.Rel(buildlogPth, path); err != nil {
					return err
				} else if err := os.Rename(path, filepath.Join(deployDir, strings.Replace(relLogPath, "/", "_", -1))); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			s.logger.Errorf("Failed to walk directory, error: %s", err)
		}
		return fmt.Errorf("Command failed, error: %s", err)
	}

	return nil
}

func (s FastlaneRunner) fastlaneDebugInfo(workDir string, useBundler bool, bundlerVersion gems.Version) (string, error) {
	factory, err := ruby.NewCommandFactory(command.NewFactory(env.NewRepository()), env.NewCommandLocator())
	if err != nil {
		return "", err
	}

	name := "fastlane"
	args := []string{"env"}
	var outBuffer bytes.Buffer
	opts := &command.Opts{
		Stdin:  strings.NewReader("n"),
		Stdout: bufio.NewWriter(&outBuffer),
		Stderr: os.Stderr,
		Dir:    workDir,
	}
	var cmd command.Command
	if useBundler {
		cmd = factory.CreateBundleExec(name, args, bundlerVersion.Version, opts)
	} else {
		cmd = factory.Create(name, args, opts)
	}

	s.logger.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", fmt.Errorf("Fastlane command (%s) failed, output: %s", cmd.PrintableCommandArgs(), outBuffer.String())
		}
		return "", fmt.Errorf("Fastlane command (%s) failed: %v", cmd.PrintableCommandArgs(), err)
	}

	return outBuffer.String(), nil
}

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

const notConnected = `Connected Apple Developer Portal Account not found.
Most likely because there is no Apple Developer Portal Account connected to the build.
Read more: https://devcenter.bitrise.io/getting-started/configuring-bitrise-steps-that-require-apple-developer-account-data/`
