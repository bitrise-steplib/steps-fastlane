package keytemplate

import (
	"bytes"
	"fmt"
	"runtime"
	"text/template"

	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

// Model ...
type Model struct {
	envRepo env.Repository
	logger  log.Logger
	os      string
	arch    string
}

type templateInventory struct {
	OS         string
	Arch       string
	Workflow   string
	Branch     string
	CommitHash string
}

// NewModel ...
func NewModel(envRepo env.Repository, logger log.Logger) Model {
	return Model{
		envRepo: envRepo,
		logger:  logger,
		os:      runtime.GOOS,
		arch:    runtime.GOARCH,
	}
}

// Evaluate returns the final string from a key template
func (m Model) Evaluate(key string) (string, error) {
	funcMap := template.FuncMap{
		"getenv":   m.getEnvVar,
		"checksum": m.checksum,
	}

	tmpl, err := template.New("").Funcs(funcMap).Parse(key)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	workflow := m.envRepo.Get("BITRISE_TRIGGERED_WORKFLOW_ID")
	branch := m.envRepo.Get("BITRISE_GIT_BRANCH")
	var commitHash = m.envRepo.Get("BITRISE_GIT_COMMIT")
	if commitHash == "" {
		commitHash = m.envRepo.Get("GIT_CLONE_COMMIT_HASH")
		m.logger.Infof("Build trigger doesn't have an explicit git commit hash, using the Git Clone Step's output for the .CommitHash template variable (value: %s)", commitHash)
	}

	inventory := templateInventory{
		OS:         m.os,
		Arch:       m.arch,
		Workflow:   workflow,
		Branch:     branch,
		CommitHash: commitHash,
	}
	m.validateInventory(inventory)

	resultBuffer := bytes.Buffer{}
	if err := tmpl.Execute(&resultBuffer, inventory); err != nil {
		return "", err
	}
	return resultBuffer.String(), nil
}

func (m Model) getEnvVar(key string) string {
	value := m.envRepo.Get(key)
	if value == "" {
		m.logger.Warnf("Environment variable %s is empty", key)
	}
	return value
}

func (m Model) validateInventory(inventory templateInventory) {
	m.warnIfEmpty("Workflow", inventory.Workflow)
	m.warnIfEmpty("Branch", inventory.Branch)
	m.warnIfEmpty("CommitHash", inventory.CommitHash)
}

func (m Model) warnIfEmpty(name, value string) {
	if value == "" {
		m.logger.Debugf("Template variable .%s is not defined", name)
	}
}
