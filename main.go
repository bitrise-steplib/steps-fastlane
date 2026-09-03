package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-steputils/v2/ruby"
	"github.com/bitrise-io/go-steputils/v2/stepconf"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/errorutil"
	. "github.com/bitrise-io/go-utils/v2/exitcode"
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/pathutil"
)

func main() {
	exitCode := run()
	os.Exit(int(exitCode))
}

func run() ExitCode {
	logger := log.NewLogger()
	buildStep, err := createStep(logger)
	if err != nil {
		logger.Println()
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to initialize Step: %w", err)))
		return Failure
	}

	config, err := buildStep.ProcessConfig()
	if err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to process Step inputs: %w", err)))
		return Failure
	}

	dependenciesOpts := EnsureDependenciesOpts{
		GemVersions:    config.GemVersions,
		UseBundler:     config.GemVersions.fastlane.Found,
		WorkDir:        config.WorkDir,
		UpdateFastlane: config.UpdateFastlane,
	}

	if err := buildStep.InstallDependencies(dependenciesOpts); err != nil {
		buildStep.logger.Println()
		buildStep.logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to install Step dependencies: %w", err)))
		return Failure
	}

	runOpts := createRunOptions(config)
	if err := buildStep.Run(runOpts); err != nil {
		buildStep.logger.Println()
		logger.Errorf(errorutil.FormattedError(fmt.Errorf("Failed to execute Step: %w", err)))
		return Failure
	}

	buildStep.tracker.wait()

	return Success
}

func createStep(logger log.Logger) (FastlaneRunner, error) {
	envRepository := env.NewRepository()
	inputParser := stepconf.NewInputParser(envRepository)
	cmdFactory := command.NewFactory(envRepository)
	cmdLocator := env.NewCommandLocator()
	// An unrecognised Ruby install type is only warned about, but a missing Ruby is fatal: the Step
	// cannot run any of its Ruby tooling without it, and carrying a nil factory around panics later.
	rbyFactory, err := ruby.NewCommandFactory(cmdFactory, cmdLocator, logger)
	if err != nil {
		return FastlaneRunner{}, err
	}

	rubyEnv := ruby.NewEnvironment(rbyFactory, cmdLocator, logger)

	pathModifier := pathutil.NewPathModifier()
	tracker := newStepTracker(envRepository, logger)

	return NewFastlaneRunner(inputParser, logger, cmdLocator, cmdFactory, rbyFactory, rubyEnv, pathModifier, tracker), nil
}

// FastlaneRunner ...
type FastlaneRunner struct {
	inputParser     stepconf.InputParser
	logger          log.Logger
	cmdFactory      command.Factory
	cmdLocator      env.CommandLocator
	rbyFactory      ruby.CommandFactory
	rubyEnvironment ruby.Environment
	pathModifier    pathutil.PathModifier
	tracker         stepTracker
}

// NewFastlaneRunner ...
func NewFastlaneRunner(
	stepInputParser stepconf.InputParser,
	logger log.Logger,
	commandLocator env.CommandLocator,
	cmdFactory command.Factory,
	rbyFactory ruby.CommandFactory,
	rubyEnvironment ruby.Environment,
	pathModifier pathutil.PathModifier,
	tracker stepTracker,
) FastlaneRunner {
	return FastlaneRunner{
		inputParser:     stepInputParser,
		logger:          logger,
		cmdLocator:      commandLocator,
		cmdFactory:      cmdFactory,
		rbyFactory:      rbyFactory,
		rubyEnvironment: rubyEnvironment,
		pathModifier:    pathModifier,
		tracker:         tracker,
	}
}

func createRunOptions(config Config) RunOpts {
	return RunOpts{
		WorkDir:         config.WorkDir,
		AuthCredentials: config.AuthCredentials,
		LaneOptions:     config.LaneOptions,
		UseBundler:      config.GemVersions.fastlane.Found,
		GemVersions:     config.GemVersions,
		EnableCache:     config.EnableCache,
	}
}
