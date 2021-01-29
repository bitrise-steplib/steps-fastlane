package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/gems"
	"github.com/bitrise-io/go-utils/command/rubycommand"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/steps-deploy-to-itunesconnect-deliver/appleauth"
	"github.com/kballard/go-shellquote"
)

// Config contains inputs parsed from environment variables
type Config struct {
	WorkDir string `env:"work_dir,dir"`
	Lane    string `env:"lane,required"`

	BitriseConnection string          `env:"connection,opt[automatic,api_key,apple_id,off]"`
	APIKeyPath        string          `env:"api_key_path"`
	APIIssuer         string          `env:"api_issuer"`
	ItunesConnectUser string          `env:"itunescon_user"`
	Password          stepconf.Secret `env:"password"`
	AppPassword       stepconf.Secret `env:"app_password"`
	TeamID            string          `env:"team_id"`
	TeamName          string          `env:"team_name"`

	UpdateFastlane bool `env:"update_fastlane,opt[true,false]"`
	VerboseLog     bool `env:"verbose_log,opt[yes,no]"`
	EnableCache    bool `env:"enable_cache,opt[yes,no]"`

	GemHome string `env:"GEM_HOME"`
}

func parseAuthSources(bitriseConnection string) ([]appleauth.Source, error) {
	switch bitriseConnection {
	case "automatic":
		return []appleauth.Source{
			&appleauth.ConnectionAPIKeySource{},
			&appleauth.ConnectionAppleIDSource{},
			&appleauth.InputAPIKeySource{},
			&appleauth.InputAppleIDSource{},
		}, nil
	case "api_key":
		return []appleauth.Source{&appleauth.ConnectionAPIKeySource{}}, nil
	case "apple_id":
		return []appleauth.Source{&appleauth.ConnectionAppleIDSource{}}, nil
	case "off":
		return []appleauth.Source{
			&appleauth.InputAPIKeySource{},
			&appleauth.InputAppleIDSource{},
		}, nil
	default:
		return nil, fmt.Errorf("invalid connection input: %s", bitriseConnection)
	}
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func fastlaneDebugInfo(workDir string, useBundler bool, bundlerVersion gems.Version) (string, error) {
	envCmd := []string{"fastlane", "env"}
	if useBundler {
		envCmd = append(gems.BundleExecPrefix(bundlerVersion), envCmd...)
	}

	cmd, err := rubycommand.NewFromSlice(envCmd)
	if err != nil {
		return "", fmt.Errorf("failed to create command model, error: %s", err)
	}

	var outBuffer bytes.Buffer
	cmd.SetStdin(strings.NewReader("n"))
	cmd.SetStdout(bufio.NewWriter(&outBuffer)).SetStderr(os.Stderr)
	cmd.SetDir(workDir)

	log.Debugf("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
	}

	return outBuffer.String(), nil
}

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		failf("Issue with input: %s", err)
	}

	stepconf.Print(config)
	log.SetEnableDebugLog(config.VerboseLog)
	fmt.Println()

	if strings.TrimSpace(config.GemHome) != "" {
		log.Warnf("Custom value (%s) is set for GEM_HOME environment variable. This can lead to errors as gem lookup path may not contain GEM_HOME.")
	}

	// Expand WorkDir
	log.Infof("Expand WorkDir")

	workDir := config.WorkDir
	if workDir == "" {
		log.Printf("WorkDir not set, using CurrentWorkingDirectory...")
		currentDir, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
		if err != nil {
			failf("Failed to get current dir, error: %s", err)
		}
		workDir = currentDir
	} else {
		absWorkDir, err := pathutil.AbsPath(workDir)
		if err != nil {
			failf("Failed to expand path (%s), error: %s", workDir, err)
		}
		workDir = absWorkDir
	}

	log.Donef("Expanded WorkDir: %s", workDir)

	if rbenvVersionsCommand := gems.RbenvVersionsCommand(); rbenvVersionsCommand != nil {
		fmt.Println()
		log.Donef("$ %s", rbenvVersionsCommand.PrintableCommandArgs())
		if err := rbenvVersionsCommand.SetStdout(os.Stdout).SetStderr(os.Stderr).SetDir(workDir).Run(); err != nil {
			log.Warnf("%s", err)
		}
	}

	//
	// Select and fetch Apple authenication source
	authSources, err := parseAuthSources(config.BitriseConnection)
	if err != nil {
		failf("Input error: unexpected value for Bitrise Apple Developer Connection (%s)", config.BitriseConnection)
	}
	authConfig, err := appleauth.Fetch(authSources, appleauth.Inputs{
		Username:            config.ItunesConnectUser,
		Password:            string(config.Password),
		AppSpecificPassword: string(config.AppPassword),
		APIIssuer:           config.APIIssuer,
		APIKeyPath:          config.APIKeyPath,
		TeamID:              config.TeamID,
		TeamName:            config.TeamName,
	})
	if err != nil {
		failf("Could not configure Apple Service authentication: %v", err)
	}

	// Split lane option
	laneOptions, err := shellquote.Split(config.Lane)
	if err != nil {
		failf("Failed to parse lane (%s), error: %s", config.Lane, err)
	}

	// Determine desired Fastlane version
	fmt.Println()
	log.Infof("Determine desired Fastlane version")

	gemVersions, err := parseGemfileLock(workDir)
	if err != nil {
		failf("%s", err)
	}

	useBundler := false
	if gemVersions.fastlane.Found {
		useBundler = true
	}

	fmt.Println()

	// Install desired Fastlane version
	if useBundler {
		log.Infof("Install bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bunder _1.2.3_" can return 'Command not found', installing bundler solves this
		installBundlerCommand := gems.InstallBundlerCommand(gemVersions.bundler)

		log.Donef("$ %s", installBundlerCommand.PrintableCommandArgs())
		fmt.Println()

		installBundlerCommand.SetStdout(os.Stdout).SetStderr(os.Stderr)
		installBundlerCommand.SetDir(workDir)

		if err := installBundlerCommand.Run(); err != nil {
			failf("command failed, error: %s", err)
		}

		// install Gemfile.lock gems with `bundle [_version_] install ...`
		fmt.Println()
		log.Infof("Install Fastlane with bundler")

		cmd, err := gems.BundleInstallCommand(gemVersions.bundler)
		if err != nil {
			failf("failed to create bundle command, error: %s", err)
		}

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(workDir)

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}
	} else if config.UpdateFastlane {
		log.Infof("Update system installed Fastlane")

		cmds, err := rubycommand.GemInstall("fastlane", "", false)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		for _, cmd := range cmds {
			log.Donef("$ %s", cmd.PrintableCommandArgs())

			cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
			cmd.SetDir(workDir)

			if err := cmd.Run(); err != nil {
				failf("Command failed, error: %s", err)
			}
		}
	} else {
		log.Infof("Using system installed Fastlane")
	}

	fmt.Println()
	log.Infof("Fastlane version")

	versionCmd := []string{"fastlane", "--version"}
	if useBundler {
		versionCmd = append(gems.BundleExecPrefix(gemVersions.bundler), versionCmd...)
	}

	log.Donef("$ %s", command.PrintableCommandArgs(false, versionCmd))

	cmd, err := rubycommand.NewFromSlice(versionCmd)
	if err != nil {
		failf("Command failed, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(workDir)

	if err := cmd.Run(); err != nil {
		failf("Command failed, error: %s", err)
	}

	// Run fastlane
	fmt.Println()
	log.Infof("Run Fastlane")

	envs := []string{}
	params, err := appleauth.AppendFastlaneCredentials(appleauth.FastlaneParams{Envs: envs, Args: laneOptions}, authConfig)
	if err != nil {
		failf("Failed to set up Apple Service authentication for Fastlane: %s", err)
	}
	envs = params.Envs
	laneOptions = append(params.Args, laneOptions...)

	fastlaneCmd := []string{"fastlane"}
	fastlaneCmd = append(fastlaneCmd, laneOptions...)
	if useBundler {
		fastlaneCmd = append(gems.BundleExecPrefix(gemVersions.bundler), fastlaneCmd...)
	}

	log.Donef("$ %s", command.PrintableCommandArgs(false, fastlaneCmd))

	cmd, err = rubycommand.NewFromSlice(fastlaneCmd)
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(workDir)

	buildlogPth := ""

	if tempDir, err := pathutil.NormalizedOSTempDirPath("fastlane_logs"); err != nil {
		log.Errorf("Failed to create temp dir for fastlane logs, error: %s", err)
	} else {
		buildlogPth = tempDir
		envs = append(envs, "FL_BUILDLOG_PATH="+buildlogPth)
	}

	cmd.AppendEnvs(envs...)

	deployDir := os.Getenv("BITRISE_DEPLOY_DIR")
	if deployDir == "" {
		log.Warnf("No BITRISE_DEPLOY_DIR found")
	}
	deployPth := filepath.Join(deployDir, "fastlane_env.log")

	if err := cmd.Run(); err != nil {
		fmt.Println()
		log.Errorf("Fastlane command: (%s) failed", cmd.PrintableCommandArgs())
		log.Errorf("If you want to send an issue report to fastlane (https://github.com/fastlane/fastlane/issues/new), you can find the output of fastlane env in the following log file:")
		fmt.Println()
		log.Infof(deployPth)
		fmt.Println()

		if fastlaneDebugInfo, err := fastlaneDebugInfo(workDir, useBundler, gemVersions.bundler); err != nil {
			log.Warnf("%s", err)
		} else if fastlaneDebugInfo != "" {
			if err := fileutil.WriteStringToFile(deployPth, fastlaneDebugInfo); err != nil {
				log.Warnf("Failed to write fastlane env log file, error: %s", err)
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
			log.Errorf("Failed to walk directory, error: %s", err)
		}
		failf("Command failed, error: %s", err)
	}

	if config.EnableCache {
		fmt.Println()
		log.Infof("Collecting cache")

		c := cache.New()
		for _, depFunc := range depsFuncs {
			includes, excludes, err := depFunc(workDir)
			log.Debugf("%s found include path:\n%s\nexclude paths:\n%s", functionName(depFunc), strings.Join(includes, "\n"), strings.Join(excludes, "\n"))
			if err != nil {
				log.Warnf("failed to collect dependencies: %s", err.Error())
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
			log.Warnf("failed to commit paths to cache: %s", err)
		}
	}
}
