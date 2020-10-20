package main

import (
	"encoding/json"
	"fmt"
	"github.com/YAWAL/mml/model"
	"io/ioutil"
	"net/http"
)

const versionManifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

func main() {

	versionManifest := model.VersionManifest{}

	resp, err := http.Get(versionManifestUrl)
	if err != err {
		fmt.Errorf("get version manifest: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != err {
		fmt.Errorf("read response body: %s", err.Error())
		return
	}
	if err := json.Unmarshal(data, &versionManifest); err != nil {
		fmt.Errorf("read response body: %s", err.Error())
		return
	}

	for _, item := range versionManifest.Versions {
		fmt.Printf("version: %s \n", item.ID)
	}

}
