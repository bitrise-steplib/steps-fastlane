package steps

const (
	ActivateSSHKeyID      = "activate-ssh-key"
	ActivateSSHKeyVersion = "4"
)

const (
	AndroidLintID      = "android-lint"
	AndroidLintVersion = "0"
)

const (
	AndroidUnitTestID      = "android-unit-test"
	AndroidUnitTestVersion = "1"
)

const (
	AndroidBuildID      = "android-build"
	AndroidBuildVersion = "1"
)

const (
	GitCloneID      = "git-clone"
	GitCloneVersion = "8"
)

const (
	CacheRestoreGradleID         = "restore-gradle-cache"
	CacheRestoreGradleVersion    = "1"
	CacheRestoreCocoapodsID      = "restore-cocoapods-cache"
	CacheRestoreCocoapodsVersion = "1"
	CacheRestoreCarthageID       = "restore-carthage-cache"
	CacheRestoreCarthageVersion  = "1"
	CacheRestoreNPMID            = "restore-npm-cache"
	CacheRestoreNPMVersion       = "1"
	CacheRestoreSPMID            = "restore-spm-cache"
	CacheRestoreSPMVersion       = "1"
	CacheRestoreDartID           = "restore-dart-cache"
	CacheRestoreDartVersion      = "1"

	CacheSaveGradleID         = "save-gradle-cache"
	CacheSaveGradleVersion    = "1"
	CacheSaveCocoapodsID      = "save-cocoapods-cache"
	CacheSaveCocoapodsVersion = "1"
	CacheSaveCarthageID       = "save-carthage-cache"
	CacheSaveCarthageVersion  = "1"
	CacheSaveNPMID            = "save-npm-cache"
	CacheSaveNPMVersion       = "1"
	CacheSaveSPMID            = "save-spm-cache"
	CacheSaveSPMVersion       = "1"
	CacheSaveDartID           = "save-dart-cache"
	CacheSaveDartVersion      = "1"
)

const (
	CertificateAndProfileInstallerID      = "certificate-and-profile-installer"
	CertificateAndProfileInstallerVersion = "1"
)

const (
	ChangeAndroidVersionCodeAndVersionNameID      = "change-android-versioncode-and-versionname"
	ChangeAndroidVersionCodeAndVersionNameVersion = "1"
)

const (
	DeployToBitriseIoID      = "deploy-to-bitrise-io"
	DeployToBitriseIoVersion = "2"
)

const (
	SignAPKID      = "sign-apk"
	SignAPKVersion = "1"
)

const (
	InstallMissingAndroidToolsID      = "install-missing-android-tools"
	InstallMissingAndroidToolsVersion = "3"
)

const (
	FastlaneID      = "fastlane"
	FastlaneVersion = "3"
)

const (
	CocoapodsInstallID      = "cocoapods-install"
	CocoapodsInstallVersion = "2"
)

const (
	CarthageID      = "carthage"
	CarthageVersion = "3"
)

const (
	XcodeArchiveID      = "xcode-archive"
	XcodeArchiveVersion = "5"
)

const (
	XcodeTestID      = "xcode-test"
	XcodeTestVersion = "5"
)

const (
	XcodeBuildForTestID      = "xcode-build-for-test"
	XcodeBuildForTestVersion = "3"
)

const (
	XcodeArchiveMacID      = "xcode-archive-mac"
	XcodeArchiveMacVersion = "1"
)

const (
	ExportXCArchiveID      = "export-xcarchive"
	ExportXCArchiveVersion = "4"
)

const (
	XcodeTestMacID      = "xcode-test-mac"
	XcodeTestMacVersion = "1"
)

const (
	CordovaArchiveID      = "cordova-archive"
	CordovaArchiveVersion = "3"
)

const (
	IonicArchiveID      = "ionic-archive"
	IonicArchiveVersion = "2"
)

const (
	GenerateCordovaBuildConfigID      = "generate-cordova-build-configuration"
	GenerateCordovaBuildConfigVersion = "0"
)

const (
	JasmineTestRunnerID      = "jasmine-runner"
	JasmineTestRunnerVersion = "0"
)

const (
	KarmaJasmineTestRunnerID      = "karma-jasmine-runner"
	KarmaJasmineTestRunnerVersion = "0"
)

const (
	NpmID      = "npm"
	NpmVersion = "1"
)

const (
	RunEASBuildID      = "run-eas-build"
	RunEASBuildVersion = "0"
)

var RunEASBuildPlatforms = []string{"all", "android", "ios"}

const (
	YarnID      = "yarn"
	YarnVersion = "0"
)

const (
	FlutterInstallID      = "flutter-installer"
	FlutterInstallVersion = "0"
)

const (
	FlutterTestID      = "flutter-test"
	FlutterTestVersion = "1"
)

const (
	FlutterAnalyzeID      = "flutter-analyze"
	FlutterAnalyzeVersion = "0"
)

const (
	FlutterBuildID      = "flutter-build"
	FlutterBuildVersion = "0"
)
