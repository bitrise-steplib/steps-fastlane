package android

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise-init/models"
)

// Scanner ...
type Scanner struct {
	SearchDir      string
	ProjectRoots   []string
	ExcludeTest    bool
	ExcludeAppIcon bool
}

// NewScanner ...
func NewScanner() *Scanner {
	return &Scanner{}
}

// Name ...
func (Scanner) Name() string {
	return ScannerName
}

// ExcludedScannerNames ...
func (*Scanner) ExcludedScannerNames() []string {
	return nil
}

// DetectPlatform ...
func (scanner *Scanner) DetectPlatform(searchDir string) (_ bool, err error) {
	scanner.SearchDir = searchDir

	scanner.ProjectRoots, err = walkMultipleFiles(searchDir, "build.gradle", "settings.gradle")
	if err != nil {
		return false, fmt.Errorf("failed to search for build.gradle files, error: %s", err)
	}

	kotlinRoots, err := walkMultipleFiles(searchDir, "build.gradle.kts", "settings.gradle.kts")
	if err != nil {
		return false, fmt.Errorf("failed to search for build.gradle files, error: %s", err)
	}

	scanner.ProjectRoots = append(scanner.ProjectRoots, kotlinRoots...)

	return len(scanner.ProjectRoots) > 0, err
}

// Options ...
func (scanner *Scanner) Options() (models.OptionNode, models.Warnings, models.Icons, error) {
	projectLocationOption := models.NewOption(ProjectLocationInputTitle, ProjectLocationInputSummary, ProjectLocationInputEnvKey, models.TypeSelector)
	warnings := models.Warnings{}
	appIconsAllProjects := models.Icons{}

	for _, projectRoot := range scanner.ProjectRoots {
		if err := checkGradlew(projectRoot); err != nil {
			return models.OptionNode{}, warnings, nil, err
		}

		relProjectRoot, err := filepath.Rel(scanner.SearchDir, projectRoot)
		if err != nil {
			return models.OptionNode{}, warnings, nil, err
		}

		icons, err := LookupIcons(projectRoot, scanner.SearchDir)
		if err != nil {
			return models.OptionNode{}, warnings, nil, err
		}
		appIconsAllProjects = append(appIconsAllProjects, icons...)
		iconIDs := make([]string, len(icons))
		for i, icon := range icons {
			iconIDs[i] = icon.Filename
		}

		configOption := models.NewConfigOption(ConfigName, iconIDs)
		moduleOption := models.NewOption(ModuleInputTitle, ModuleInputSummary, ModuleInputEnvKey, models.TypeUserInput)
		variantOption := models.NewOption(VariantInputTitle, VariantInputSummary, VariantInputEnvKey, models.TypeOptionalUserInput)

		projectLocationOption.AddOption(relProjectRoot, moduleOption)
		moduleOption.AddOption("app", variantOption)
		variantOption.AddConfig("", configOption)
	}

	return *projectLocationOption, warnings, appIconsAllProjects, nil
}

// DefaultOptions ...
func (scanner *Scanner) DefaultOptions() models.OptionNode {
	projectLocationOption := models.NewOption(ProjectLocationInputTitle, ProjectLocationInputSummary, ProjectLocationInputEnvKey, models.TypeUserInput)
	moduleOption := models.NewOption(ModuleInputTitle, ModuleInputSummary, ModuleInputEnvKey, models.TypeUserInput)
	variantOption := models.NewOption(VariantInputTitle, VariantInputSummary, VariantInputEnvKey, models.TypeOptionalUserInput)
	configOption := models.NewConfigOption(DefaultConfigName, nil)

	projectLocationOption.AddOption("", moduleOption)
	moduleOption.AddOption("", variantOption)
	variantOption.AddConfig("", configOption)

	return *projectLocationOption
}

// Configs ...
func (scanner *Scanner) Configs() (models.BitriseConfigMap, error) {
	configBuilder := scanner.generateConfigBuilder()

	config, err := configBuilder.Generate(ScannerName)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	return models.BitriseConfigMap{
		ConfigName: string(data),
	}, nil
}

// DefaultConfigs ...
func (scanner *Scanner) DefaultConfigs() (models.BitriseConfigMap, error) {
	configBuilder := scanner.generateConfigBuilder()

	config, err := configBuilder.Generate(ScannerName)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	return models.BitriseConfigMap{
		DefaultConfigName: string(data),
	}, nil
}
