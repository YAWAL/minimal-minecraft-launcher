package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/YAWAL/mml/model"
)

const (
	versionManifestUrl    = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	minecraftResourcesUrl = "https://resources.download.minecraft.net"
	minecraftPath         = "temp"         // TODO: user should set this
	username              = "playername"   // TODO: user should set this
	minecraftVersion      = "1.16.2"       // TODO: user should set this
	accessToken           = "youracctoken" // TODO: user should set this
)

func assetsPath() string {
	return minecraftPath + "/assets/"
}

func clientPath(versionDetails *model.VersionDetails) string {
	return minecraftPath + "/versions/" + versionDetails.ID + "/" + versionDetails.ID + ".jar"
}

func main() {
	now := time.Now()
	versionManifest := model.VersionManifest{}

	if err := doRequest(versionManifestUrl, &versionManifest); err != nil {
		fmt.Printf("get version manifest: %s", err.Error())
		return
	}

	//for _, item := range versionManifest.Versions {
	//	fmt.Printf("version: %s \n", item.ID)
	//}

	var version model.Version

	for _, item := range versionManifest.Versions {
		if item.ID == minecraftVersion {
			version = item
		}
	}

	versionDetails := model.VersionDetails{}
	if err := doRequest(version.URL, &versionDetails); err != nil {
		fmt.Printf("get version details: %s", err.Error())
		return
	}

	if err := downloadLibraries(versionDetails.Libraries); err != nil {
		fmt.Printf("get libraries: %s", err.Error())
		return
	}

	if err := downloadClient(&versionDetails); err != nil {
		fmt.Printf("get client: %s", err.Error())
		return
	}

	assets := model.AssetsData{}
	if err := doRequest(versionDetails.AssetIndex.URL, &assets); err != nil {
		fmt.Printf("get assets: %s", err.Error())
		return
	}

	if err := downloadIndexJson(&versionDetails.AssetIndex); err != nil {
		fmt.Printf("get assets index json: %s", err.Error())
		return
	}

	if err := downloadResources(&assets); err != nil {
		fmt.Printf("get resources: %s", err.Error())
		return
	}

	if err := createExecutableFile(&versionDetails); err != nil {
		fmt.Printf("create executable file: %s", err.Error())
		return
	}
	fmt.Printf("exec time: %f ", time.Since(now).Seconds())
}

func doRequest(url string, out interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &out)
}

func downloadLibraries(libraries []model.Library) error {
	libPath := minecraftPath + "/libraries/"
	for _, lib := range libraries {
		err := download(lib.Downloads.Artifact.URL, libPath+lib.Downloads.Artifact.Path)
		if err != nil {
			return err
		}
		classifiers := lib.Downloads.Classifiers
		if classifiers != nil {
			switch runtime.GOOS {
			case "linux":
				err = download(classifiers.NativesLinux.URL, libPath+classifiers.NativesLinux.Path)
			case "windows":
				err = download(classifiers.NativesWindows.URL, libPath+classifiers.NativesWindows.Path)
			case "darwin":
				err = download(classifiers.NativesMacos.URL, libPath+classifiers.NativesMacos.Path)
			default:
				err = errors.New("download libraries: unsupported OS")
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createClassPath(details *model.VersionDetails) (string, error) {
	var separator, classPath string
	var err error
	switch runtime.GOOS {
	case "linux", "darwin":
		separator = ":"
	case "windows":
		separator = ";"
	default:
		return "", errors.New("choosing separator: unsupported OS")
	}
	libPath := minecraftPath + "/libraries/"
	for _, lib := range details.Libraries {
		classPath += filepath.Clean(libPath+lib.Downloads.Artifact.Path) + separator
		classifiers := lib.Downloads.Classifiers
		if classifiers != nil {
			switch runtime.GOOS {
			case "linux":
				classPath += filepath.Clean(libPath+classifiers.NativesLinux.Path) + separator
			case "windows":
				classPath += filepath.Clean(libPath+classifiers.NativesWindows.Path) + separator
			case "darwin":
				classPath += filepath.Clean(libPath+classifiers.NativesMacos.Path) + separator
			default:
				return "", errors.New("creating classPath: unsupported OS")
			}
		}
	}
	classPath += filepath.Clean(clientPath(details))
	return classPath, err
}

func downloadResources(assets *model.AssetsData) error {
	objectsPath := assetsPath() + "objects/"

	for _, val := range assets.Objects {
		url := minecraftResourcesUrl + "/" + (val.Hash)[0:2] + "/" + val.Hash
		fullPath := objectsPath + (val.Hash)[0:2] + "/" + val.Hash
		err := download(url, fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadIndexJson(assetIndex *model.AssetIndex) error {
	indexesPath := assetsPath() + "indexes/"
	return download(assetIndex.URL, indexesPath+assetIndex.ID+".json")
}

func downloadClient(versionDetails *model.VersionDetails) error {
	return download(versionDetails.Downloads.Client.URL, clientPath(versionDetails))
}

func download(url, fullPath string) error {
	_, err := os.Stat(fullPath)
	if url == "" || err == nil {
		return nil
	}
	folder := path.Dir(fullPath)
	if err := os.MkdirAll(folder, 0700); err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(rawData)
	if err != nil {
		return err
	}

	return nil
}

func runMinecraft(args string) error {
	return exec.Command("java", args).Run()
}

func createExecutableFile(versionDetails *model.VersionDetails) error {
	classPath, err := createClassPath(versionDetails)
	if err != nil {
		return err
	}
	var fileName, prefix string
	switch runtime.GOOS {
	case "linux", "darwin":
		fileName = "start.sh"
		prefix = "#! /bin/sh \n"
	case "windows":
		fileName = "start.bat"
	default:
		return errors.New("can't create executable file: unsupported OS")
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(prefix +
		"java " +
		"-cp " + classPath + " " + versionDetails.MainClass + " " +
		"--username " + username + " " +
		"--gameDir " + minecraftPath + " " +
		"--assetIndex " + versionDetails.AssetIndex.ID + " " +
		"--assetsDir " + minecraftPath + "/assets/" + " " +
		"--accessToken " + accessToken + " " +
		"--version " + versionDetails.ID))

	defer file.Close()
	return err
}
