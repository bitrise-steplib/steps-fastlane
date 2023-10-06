package steps

import (
	"github.com/bitrise-io/bitrise-init/models"
	bitriseModels "github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/pointers"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// PrepareListParams describes the default prepare Step options.
type PrepareListParams struct {
	SSHKeyActivation models.SSHKeyActivation
}

func stepIDComposite(ID, version string) string {
	if version != "" {
		return ID + "@" + version
	}
	return ID
}

func stepListItem(stepIDComposite, title, runIf string, inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	step := stepmanModels.StepModel{}
	if title != "" {
		step.Title = pointers.NewStringPtr(title)
	}
	if runIf != "" {
		step.RunIf = pointers.NewStringPtr(runIf)
	}
	if len(inputs) > 0 {
		step.Inputs = inputs
	}

	return bitriseModels.StepListItemModel{
		stepIDComposite: step,
	}
}

func DefaultPrepareStepList(params PrepareListParams) []bitriseModels.StepListItemModel {
	stepList := []bitriseModels.StepListItemModel{}

	switch params.SSHKeyActivation {
	case models.SSHKeyActivationNone:
		{
			// No SSH key setup needed
		}
	case models.SSHKeyActivationMandatory:
		{
			// This needs the `SSH_RSA_PRIVATE_KEY` env to be defined, which depends on the selected path in the website
			// add-new-app flow.
			stepList = append(stepList, ActivateSSHKeyStepListItem(""))
		}
	case models.SSHKeyActivationConditional:
		{
			// Add the SSH key step just in case, but only run it if the required env var is available.
			runCondition := `{{getenv "SSH_RSA_PRIVATE_KEY" | ne ""}}`
			stepList = append(stepList, ActivateSSHKeyStepListItem(runCondition))
		}
	}

	stepList = append(stepList, GitCloneStepListItem())

	return stepList
}

func DefaultDeployStepList() []bitriseModels.StepListItemModel {
	stepList := []bitriseModels.StepListItemModel{}

	stepList = append(stepList, DeployToBitriseIoStepListItem())

	return stepList
}

func ActivateSSHKeyStepListItem(runIfCondition string) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(ActivateSSHKeyID, ActivateSSHKeyVersion)
	return stepListItem(stepIDComposite, "", runIfCondition)
}

func AndroidLintStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(AndroidLintID, AndroidLintVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func AndroidUnitTestStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(AndroidUnitTestID, AndroidUnitTestVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func AndroidBuildStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(AndroidBuildID, AndroidBuildVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func GitCloneStepListItem() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(GitCloneID, GitCloneVersion)
	return stepListItem(stepIDComposite, "", "")
}

func CertificateAndProfileInstallerStepListItem() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CertificateAndProfileInstallerID, CertificateAndProfileInstallerVersion)
	return stepListItem(stepIDComposite, "", "")
}

func ChangeAndroidVersionCodeAndVersionNameStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(ChangeAndroidVersionCodeAndVersionNameID, ChangeAndroidVersionCodeAndVersionNameVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func DeployToBitriseIoStepListItem() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(DeployToBitriseIoID, DeployToBitriseIoVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SignAPKStepListItem() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(SignAPKID, SignAPKVersion)
	return stepListItem(stepIDComposite, "", `{{getenv "BITRISEIO_ANDROID_KEYSTORE_URL" | ne ""}}`)
}

func InstallMissingAndroidToolsStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(InstallMissingAndroidToolsID, InstallMissingAndroidToolsVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func FastlaneStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(FastlaneID, FastlaneVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func CocoapodsInstallStepListItem() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CocoapodsInstallID, CocoapodsInstallVersion)
	return stepListItem(stepIDComposite, "", "", envmanModels.EnvironmentItemModel{
		"is_cache_disabled": "true", // Disable legacy caching when used in workflows with KV caching
	})
}

func CarthageStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CarthageID, CarthageVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func XcodeArchiveStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(XcodeArchiveID, XcodeArchiveVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func XcodeBuildForTestStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(XcodeBuildForTestID, XcodeBuildForTestVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func XcodeTestStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(XcodeTestID, XcodeTestVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func XcodeArchiveMacStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(XcodeArchiveMacID, XcodeArchiveMacVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func ExportXCArchiveStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(ExportXCArchiveID, ExportXCArchiveVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func XcodeTestMacStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(XcodeTestMacID, XcodeTestMacVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func CordovaArchiveStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CordovaArchiveID, CordovaArchiveVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func IonicArchiveStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(IonicArchiveID, IonicArchiveVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func GenerateCordovaBuildConfigStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(GenerateCordovaBuildConfigID, GenerateCordovaBuildConfigVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func JasmineTestRunnerStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(JasmineTestRunnerID, JasmineTestRunnerVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func KarmaJasmineTestRunnerStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(KarmaJasmineTestRunnerID, KarmaJasmineTestRunnerVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func NpmStepListItem(command, workdir string) bitriseModels.StepListItemModel {
	var inputs []envmanModels.EnvironmentItemModel
	if workdir != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{"workdir": workdir})
	}
	if command != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{"command": command})
	}

	stepIDComposite := stepIDComposite(NpmID, NpmVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func RunEASBuildStepListItem(workdir, platform string) bitriseModels.StepListItemModel {
	var inputs []envmanModels.EnvironmentItemModel
	if platform != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{"platform": platform})
	}
	if workdir != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{"work_dir": workdir})
	}
	stepIDComposite := stepIDComposite(RunEASBuildID, RunEASBuildVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func YarnStepListItem(command, workdir string) bitriseModels.StepListItemModel {
	var inputs []envmanModels.EnvironmentItemModel
	if workdir != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{"workdir": workdir})
	}
	if command != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{"command": command})
	}

	stepIDComposite := stepIDComposite(YarnID, YarnVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func FlutterInstallStepListItem(version string, isUpdate bool) bitriseModels.StepListItemModel {
	const versionInputKey = "version"
	const isUpdateInputKey = "is_update"

	var inputs []envmanModels.EnvironmentItemModel

	if version != "" {
		inputs = append(inputs, envmanModels.EnvironmentItemModel{versionInputKey: version})
	}

	isUpdateStr := "false"
	if isUpdate {
		isUpdateStr = "true"
	}

	inputs = append(inputs, envmanModels.EnvironmentItemModel{isUpdateInputKey: isUpdateStr})

	stepIDComposite := stepIDComposite(FlutterInstallID, FlutterInstallVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func FlutterTestStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(FlutterTestID, FlutterTestVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func FlutterAnalyzeStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(FlutterAnalyzeID, FlutterAnalyzeVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}

func FlutterBuildStepListItem(inputs ...envmanModels.EnvironmentItemModel) bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(FlutterBuildID, FlutterBuildVersion)
	return stepListItem(stepIDComposite, "", "", inputs...)
}
