package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-steputils/ruby"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/env"
	"github.com/bitrise-io/go-utils/log"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := log.NewLogger()
	buildStep := createStep(logger)

	config, err := buildStep.ProcessConfig()
	if err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(formattedError(fmt.Errorf("Failed to process Step inputs: %w", err)))
		return 1
	}

	// Determine desired Fastlane version
	fmt.Println()
	buildStep.logger.Infof("Determine desired Fastlane version")
	gemVersions, err := parseGemfileLock(config.WorkDir)
	if err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(err.Error())
	}

	fmt.Println()

	dependenciesOpts := EnsureDependenciesOpts{
		GemVersions: gemVersions,
		UseBundler:  gemVersions.fastlane.Found,
	}
	if err = buildStep.InstallDependencies(config, dependenciesOpts); err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(formattedError(fmt.Errorf("Failed to install Step dependencies: %w", err)))
		return 1
	}

	err = buildStep.Run(config, dependenciesOpts)
	if err != nil {
		buildStep.logger.Println()
		logger.Errorf(formattedError(fmt.Errorf("Failed to execute Step main logic: %w", err)))
		return 1
	}

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

// NewFastlaneRunner ...
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

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
