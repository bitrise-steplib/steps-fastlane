package android

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"github.com/bitrise-io/bitrise-init/analytics"
	"github.com/bitrise-io/bitrise-init/models"
	"github.com/bitrise-io/bitrise-init/utility"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
)

type icon struct {
	prefix       string
	fileNameBase string
}

func lookupIconName(manifestPth string) ([]icon, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(manifestPth); err != nil {
		return nil, err
	}

	log.Debugf("Looking for app icons. Manifest path: %s", manifestPth)
	return parseIconName(doc)
}

// parseIconName fetches icon name from AndroidManifest.xml.
func parseIconName(doc *etree.Document) ([]icon, error) {
	man := doc.SelectElement("manifest")
	if man == nil {
		return nil, fmt.Errorf("key 'manifest' not found in AndroidManifest.xml")
	}
	app := man.SelectElement("application")
	if app == nil {
		return nil, fmt.Errorf("key 'application' not found in AndroidManifest.xml")
	}
	ic := app.SelectAttr("android:icon")
	if ic == nil {
		// Gradle varaibles like ${appIcon} are not supported
		return nil, fmt.Errorf("attribute 'android:icon' not found in AndroidManifest.xml")
	}

	iconPathParts := strings.Split(strings.TrimPrefix(ic.Value, "@"), "/")
	if len(iconPathParts) != 2 {
		return nil, fmt.Errorf("unsupported icon key (%s)", ic.Value)
	}
	return []icon{{
		prefix:       iconPathParts[0],
		fileNameBase: iconPathParts[1],
	}}, nil
}

func lookupIconPaths(resPth string, icon icon) ([]string, error) {
	var resourceSuffixes = [...]string{"xxxhdpi", "xxhdpi", "xhdpi", "hdpi", "mdpi", "ldpi"}
	resourceDirs := make([]string, len(resourceSuffixes))
	for _, mipmapSuffix := range resourceSuffixes {
		resourceDirs = append(resourceDirs, icon.prefix+"-"+mipmapSuffix)
	}

	for _, dir := range resourceDirs {
		iconPaths, err := filepath.Glob(filepath.Join(regexp.QuoteMeta(resPth), dir, icon.fileNameBase+".png"))
		if err != nil {
			return nil, err
		}
		if len(iconPaths) != 0 {
			return iconPaths, nil
		}
	}
	return nil, nil
}

func lookupIcons(projectDir string, basepath string) ([]string, error) {
	variantPaths := filepath.Join(regexp.QuoteMeta(projectDir), "*", "src", "*")
	manifestPaths, err := filepath.Glob(filepath.Join(variantPaths, "AndroidManifest.xml"))
	if err != nil {
		return nil, err
	}
	resourcesPaths, err := filepath.Glob(filepath.Join(variantPaths, "res"))
	if err != nil {
		return nil, err
	}

	// falling back to standard icon name, if not found in manifest
	iconNames := []icon{
		{
			prefix:       "mipmap",
			fileNameBase: "ic_launcher",
		},
		{
			prefix:       "mipmap",
			fileNameBase: "ic_launcher_round",
		},
	}
	for _, manifestPath := range manifestPaths {
		icons, err := lookupIconName(manifestPath)
		if err != nil {
			analytics.LogInfo("android-icon-lookup", analytics.DetectorErrorData("android", err), "Failed to lookup android icon")
			continue
		}

		iconNames = append(iconNames, icons...)
	}

	var iconPaths []string
	for _, resourcesPath := range resourcesPaths {
		for _, icon := range iconNames {
			foundIconPaths, err := lookupIconPaths(resourcesPath, icon)
			if err != nil {
				return nil, err
			}

			iconPaths = append(iconPaths, foundIconPaths...)
		}
	}
	return sliceutil.UniqueStringSlice(iconPaths), nil
}

// LookupIcons returns the largest resolution for all potential android icons.
func LookupIcons(projectDir string, basepath string) (models.Icons, error) {
	iconPaths, err := lookupIcons(projectDir, basepath)
	if err != nil {
		return nil, err
	}
	return utility.CreateIconDescriptors(iconPaths, basepath)
}
