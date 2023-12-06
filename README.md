# fastlane

[![Step changelog](https://shields.io/github/v/release/bitrise-io/steps-fastlane?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-io/steps-fastlane/releases)

Runs your fastlane lane on [bitrise.io](https://www.bitrise.io).

<details>
<summary>Description</summary>

fastlane is an open-source app automation tool for iOS, Android and for most cross-platform frameworks, for example, React Native, Xamarin and Flutter.
**fastlane** Step helps you integrate your lane to your Bitrise Workflow and runs your lane based on the fastlane actions with minimal configuration.
If your Apple Developer Portal account is [connected to Bitrise](https://devcenter.bitrise.io/getting-started/connecting-to-services/configuring-bitrise-steps-that-require-apple-developer-account-data/), the `FASTLANE_SESSION` Environment Variable will pass on the session data to fastlane.

### Configuring the Step

Before you start configuring the Step, make sure you've [connected to Apple services either by API key, Apple ID or through Fastlane Step's input fields](https://devcenter.bitrise.io/getting-started/connecting-to-services/bitrise-steps-and-their-authentication-methods/#fastlane-step).
1. Add the **fastlane** Step to your Workflow after the **Git Clone Repository** Step or any other dependency Step.
1. Based on your project's fastlane setup, you can add your project's default lane or a custom lane in the **fastlane lane** input.
2. If your fastlane directory is not available in your repository's root, then you can add the parent directory of fastlane directory in the **Working directory** input.
3. If your project doesn't contain a fastlane gem in your project's Gemfile, you can use the **Should update fastlane gem before run** input.
Set this input to `true` so that the Step can install the latest fastlane version to your project.
If a gem lockfile (Gemfile.lock or gems.locked) includes the fastlane gem in the working directory, that specific fastlane version will be installed.
4. Select `yes` in the **Enable verbose logging** input if you wish to run your build in debug mode and print out error additional debug logs.
5. Select `yes` in the **Enable collecting files to be included in the build cache** to cache pods, Carthage and Android dependencies.

### Troubleshooting
If you run your lane on Bitrise and your build fails on the **fastlane** Step, the logs won't reveal too much about the error since it's most likely related to the fastlane file's configuration.
We recommend you swap your fastlane actions for Bitrise Steps which will bring out the problem.

### Useful links
- [About fastlane](https://docs.fastlane.tools)
- [Connecting your Apple Developer Account to Bitrise](https://devcenter.bitrise.io/getting-started/connecting-to-services/configuring-bitrise-steps-that-require-apple-developer-account-data/)
- [Running fastlane on Bitrise](https://devcenter.bitrise.io/tutorials/fastlane/fastlane-index/)

### Related Steps
- [Deploy to iTunes Connect/Deliver](https://www.bitrise.io/integrations/steps/deploy-to-itunesconnect-deliver)
- [iOS Auto Provision](https://www.bitrise.io/integrations/steps/ios-auto-provision)
- [Fastlane Match](https://www.bitrise.io/integrations/steps/fastlane-match)
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/steps-and-workflows/steps-and-workflows-index/).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

### Examples

Run the `tests` lane from the current dir:
```yaml
- fastlane:
    inputs:
    - lane: tests
    - work_dir: ./
```


## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `lane` | fastlane lane to run $ fastlane [lane]  | required |  |
| `work_dir` | Use this option if the fastlane directory is not in your repository's root.  Working directory should be the parent directory of your Fastfile's directory.  For example:  * If the Fastfile path is `./here/is/my/fastlane/Fastfile` * Then the Fastfile's directory is `./here/is/my/fastlane` * So the Working Directory should be `./here/is/my` |  | `$BITRISE_SOURCE_DIR` |
| `connection` | The input determines the method used for Apple Service authentication. By default, any enabled Bitrise Apple Developer connection is used and other authentication-related Step inputs are ignored.  There are two types of Apple Developer connection you can enable on Bitrise: one is based on an API key of the App Store Connect API, the other is the session-based authentication with an Apple ID. You can choose which type of Bitrise Apple Developer connection to use or you can tell the Step to only use the Step inputs for authentication: - `automatic`: Use any enabled Apple Developer connection, either based on Apple ID authentication or API key authentication.  Step inputs are only used as a fallback. API key authentication has priority over Apple ID authentication in both cases. - `api_key`: Use the Apple Developer connection based on API key authentication. Authentication-related Step inputs are ignored. - `apple_id`: Use the Apple Developer connection based on Apple ID authentication and the **Application-specific password** Step input. Other authentication-related Step inputs are ignored. - `off`: Do not use any already configured Apple Developer Connection. Only authentication-related Step inputs are considered. | required | `automatic` |
| `api_key_path` | Specify the path in an URL format where your API key is stored. For example: `https://URL/TO/AuthKey_[KEY_ID].p8` or `file:///PATH/TO/AuthKey_[KEY_ID].p8`. **NOTE:** The Step will only recognize the API key if the filename includes the  `KEY_ID` value as shown on the examples above.  You can upload your key on the **Generic File Storage** tab in the Workflow Editor and set the Environment Variable for the file here.  For example: `$BITRISEIO_MYKEY_URL` |  |  |
| `api_issuer` | Issuer ID. Required if **API Key: URL** (`api_key_path`) is specified. |  |  |
| `apple_id` | Email for Apple ID login. | sensitive |  |
| `password` | Password for the specified Apple ID. | sensitive |  |
| `app_password` | Use this input if TFA is enabled on the Apple ID but no app-specific password has been added to the used Bitrise Apple ID connection.  **NOTE:** Application-specific passwords can be created on the [AppleID Website](https://appleid.apple.com). It can be used to bypass two-factor authentication. | sensitive |  |
| `update_fastlane` | Should update fastlane gem before run? *This option will be skipped if you have a `Gemfile` in the `work_dir` directory.* |  | `true` |
| `verbose_log` | Enable/disable verbose logging. | required | `no` |
| `enable_cache` | If enabled the step will add the following cache items (if they exist): - Pods -> Podfile.lock - Carthage -> Cartfile.resolved - Android dependencies | required | `yes` |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-io/steps-fastlane/pulls) and [issues](https://github.com/bitrise-io/steps-fastlane/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)
