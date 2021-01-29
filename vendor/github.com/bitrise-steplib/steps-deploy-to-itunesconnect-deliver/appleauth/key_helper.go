package appleauth

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bitrise-io/go-utils/log"
)

func fetchPrivateKey(privateKeyURL string) ([]byte, string, error) {
	fileURL, err := url.Parse(privateKeyURL)
	if err != nil {
		return nil, "", err
	}

	keyID := getKeyID(fileURL)
	// .appstoreconnect/private_keys is a  path searched by altool (see altool's man page)
	keyFile := filepath.Join(os.Getenv("HOME"), ".appstoreconnect/private_keys", fmt.Sprintf("AuthKey_%s.p8", keyID))
	if err := copyOrDownloadFile(fileURL, keyFile); err != nil {
		return nil, "", err
	}

	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, "", err
	}

	return key, keyID, nil
}

func copyOrDownloadFile(u *url.URL, pth string) error {
	if err := os.MkdirAll(filepath.Dir(pth), 0777); err != nil {
		return err
	}

	certFile, err := os.Create(pth)
	if err != nil {
		return err
	}
	defer func() {
		if err := certFile.Close(); err != nil {
			log.Errorf("Failed to close file, error: %s", err)
		}
	}()

	// if file -> copy
	if u.Scheme == "file" {
		b, err := ioutil.ReadFile(u.Path)
		if err != nil {
			return err
		}
		_, err = certFile.Write(b)
		return err
	}

	// otherwise download
	f, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Body.Close(); err != nil {
			log.Errorf("Failed to close file, error: %s", err)
		}
	}()

	_, err = io.Copy(certFile, f.Body)
	return err
}

func getKeyID(u *url.URL) string {
	var keyID = "Bitrise" // as default if no ID found in file name

	// get the ID of the key from the file
	if matches := regexp.MustCompile(`AuthKey_(.+)\.p8`).FindStringSubmatch(filepath.Base(u.Path)); len(matches) == 2 {
		keyID = matches[1]
	}

	return keyID
}
