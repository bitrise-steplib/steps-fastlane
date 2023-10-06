package steps

import bitriseModels "github.com/bitrise-io/bitrise/models"

func RestoreGradleCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheRestoreGradleID, CacheRestoreGradleVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SaveGradleCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheSaveGradleID, CacheSaveGradleVersion)
	return stepListItem(stepIDComposite, "", "")
}

func RestoreCocoapodsCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheRestoreCocoapodsID, CacheRestoreCocoapodsVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SaveCocoapodsCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheSaveCocoapodsID, CacheSaveCocoapodsVersion)
	return stepListItem(stepIDComposite, "", "")
}

func RestoreCarthageCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheRestoreCarthageID, CacheRestoreCarthageVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SaveCarthageCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheSaveCarthageID, CacheSaveCarthageVersion)
	return stepListItem(stepIDComposite, "", "")
}

func RestoreNPMCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheRestoreNPMID, CacheRestoreNPMVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SaveNPMCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheSaveNPMID, CacheSaveNPMVersion)
	return stepListItem(stepIDComposite, "", "")
}

func RestoreSPMCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheRestoreSPMID, CacheRestoreSPMVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SaveSPMCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheSaveSPMID, CacheSaveSPMVersion)
	return stepListItem(stepIDComposite, "", "")
}

func RestoreDartCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheRestoreDartID, CacheRestoreDartVersion)
	return stepListItem(stepIDComposite, "", "")
}

func SaveDartCache() bitriseModels.StepListItemModel {
	stepIDComposite := stepIDComposite(CacheSaveDartID, CacheSaveDartVersion)
	return stepListItem(stepIDComposite, "", "")
}
