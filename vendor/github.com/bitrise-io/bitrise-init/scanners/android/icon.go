package android

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beevik/etree"
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
	parsedIcons, err := parseIconName(doc)
	if err != nil {
		return nil, err
	}

	return parsedIcons, nil
}

// parseIconName fetches icon name from AndroidManifest.xml.
func parseIconName(doc *etree.Document) ([]icon, error) {
	man := doc.SelectElement("manifest")
	if man == nil {
		log.Debugf("Key manifest not found in manifest file")
		return nil, nil
	}
	app := man.SelectElement("application")
	if app == nil {
		log.Debugf("Key application not found in manifest file")
		return nil, nil
	}
	ic := app.SelectAttr("android:icon")
	if ic == nil {
		log.Debugf("Attribute not found in manifest file")
		return nil, nil
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
			return nil, err
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
