package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/YAWAL/mml/model"
)

const (
	versionManifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	minecraftPath      = "" // TODO: user should set this
)

func main() {

	versionManifest := model.VersionManifest{}

	if err := doRequest(versionManifestUrl, &versionManifest); err != err {
		fmt.Errorf("get version manifest: %s", err.Error())
		return
	}

	//for _, item := range versionManifest.Versions {
	//	fmt.Printf("version: %s \n", item.ID)
	//}

	// TODO: refactor this
	vrs := "1.16.2"

	var version model.Version

	for _, item := range versionManifest.Versions {
		if item.ID == vrs {
			version = item
		}
	}

	versionDetails := model.VersionDetails{}

	if err := doRequest(version.URL, &versionDetails); err != err {
		fmt.Errorf("get version details: %s", err.Error())
		return
	}

	for _, item := range versionDetails.Libraries {
		fmt.Printf("version: %s \n", item.Downloads.Artifact.URL)
		fmt.Printf("version: %s \n", item.Downloads.Artifact.Path)
	}
	err := downloadLibraries(versionDetails.Libraries)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func doRequest(url string, out interface{}) error {
	resp, err := http.Get(url)
	if err != err {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != err {
		return err
	}
	return json.Unmarshal(data, &out)
}

func downloadLibraries(libraries []model.Library) error {
	libPath := minecraftPath + "libraries/"

	for _, lib := range libraries {
		fullPath := libPath + lib.Downloads.Artifact.Path
		folder := path.Dir(fullPath)

		err := os.MkdirAll(folder, 777)
		if err != err {
			return err
		}
		resp, err := http.Get(lib.Downloads.Artifact.URL)
		if err != err {
			return err
		}
		defer resp.Body.Close()

		rawData, err := ioutil.ReadAll(resp.Body)
		if err != err {
			return err
		}
		err = ioutil.WriteFile(fullPath, rawData, 777)
		if err != err {
			return err
		}

	}
	return nil
}
