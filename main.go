package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"

	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
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

	return 0
}

func createStep(logger log.Logger) FastlaneRunner {
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(env.NewRepository())
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
