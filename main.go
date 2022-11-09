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
	"github.com/bitrise-io/go-utils/v2/pathutil"
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

	dependenciesOpts := EnsureDependenciesOpts{
		GemVersions:    config.GemVersions,
		UseBundler:     config.GemVersions.fastlane.Found,
		WorkDir:        config.WorkDir,
		UpdateFastlane: config.UpdateFastlane,
	}

	if err = buildStep.InstallDependencies(dependenciesOpts); err != nil {
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

	buildStep.cacheDeps(config)

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

	pathModifier := pathutil.NewPathModifier()

	return NewFastlaneRunner(inputParser, logger, cmdLocator, cmdFactory, rbyFactory, pathModifier)
}

// FastlaneRunner ...
type FastlaneRunner struct {
	inputParser  stepconf.InputParser
	logger       log.Logger
	cmdFactory   command.Factory
	cmdLocator   env.CommandLocator
	rbyFactory   ruby.CommandFactory
	pathModifier pathutil.PathModifier
}

// NewFastlaneRunner ...
func NewFastlaneRunner(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	commandLocator env.CommandLocator,
	cmdFactory command.Factory,
	rbyFactory ruby.CommandFactory,
	pathModifier pathutil.PathModifier,
) FastlaneRunner {
	return FastlaneRunner{
		inputParser:  stepInputParser,
		logger:       logger,
		cmdLocator:   commandLocator,
		cmdFactory:   cmdFactory,
		rbyFactory:   rbyFactory,
		pathModifier: pathModifier,
	}
}

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
