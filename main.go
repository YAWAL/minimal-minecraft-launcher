package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

const (
	versionManifestUrl    = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	minecraftResourcesUrl = "https://resources.download.minecraft.net"
)

var minecraftPath, username, minecraftVersion, accessToken, initialHeapSize, maxHeapSize string

func assetsPath() string {
	return minecraftPath + "/assets/"
}

func clientPath(versionDetails *VersionDetails) string {
	return minecraftPath + "/versions/" + versionDetails.ID + "/" + versionDetails.ID + ".jar"
}

func main() {
	now := time.Now()

	if !getUserConfiguration() {
		return
	}

	versionManifest := VersionManifest{}

	if err := doRequest(versionManifestUrl, &versionManifest); err != nil {
		fmt.Printf("get version manifest: %s", err.Error())
		return
	}

	version, err := getVersion(versionManifest.Versions)
	if err != nil {
		fmt.Printf("get version: %s", err)
		return
	}

	versionDetails := VersionDetails{}
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

	assets := AssetsData{}
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

func getUserConfiguration() bool {
	flag.StringVar(&minecraftPath, "path", "temp", "Path to folder, where the game should be installed.")
	flag.StringVar(&username, "username", "playername", "Your username.")
	flag.StringVar(&minecraftVersion, "version", "1.16.1", "Select version of minecraft.") // TODO: change this to latest version
	flag.StringVar(&accessToken, "token", "youracctoken", "Your access token from Mojang, if you have this.")
	flag.StringVar(&initialHeapSize, "init_memory", "512", "Memory (in megabytes) for java machine at start.")
	flag.StringVar(&maxHeapSize, "max_memory", "2048", "Maximum allowed memory (in megabytes) for java machine.")
	flag.Parse()
	if len(os.Args) < 2 {
		flag.PrintDefaults()
		return false
	}
	return true
}

func getVersion(versions []Version) (Version, error) {
	for _, item := range versions {
		if item.ID == minecraftVersion {
			return item, nil
		}
	}
	return Version{}, fmt.Errorf("can't find version %s", minecraftVersion)
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

func downloadLibraries(libraries []Library) error {
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

func createClassPath(details *VersionDetails) (string, error) {
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

func downloadResources(assets *AssetsData) error {
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

func downloadIndexJson(assetIndex *AssetIndex) error {
	indexesPath := assetsPath() + "indexes/"
	return download(assetIndex.URL, indexesPath+assetIndex.ID+".json")
}

func downloadClient(versionDetails *VersionDetails) error {
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

func createExecutableFile(versionDetails *VersionDetails) error {
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
	// TODO: maybe put this to minecraft folder, but then we need to change path for classpath,
	// to save ability launch from relative path
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(prefix +
		"java " +
		//"-Djava.library.path=" + filepath.Clean(minecraftPath+"/versions/"+versionDetails.ID+"/natives") + " " +
		"-Xms" + initialHeapSize + "m " +
		"-Xmx" + maxHeapSize + "m " +
		"-cp " + classPath + " " + versionDetails.MainClass + " " +
		"--username " + username + " " +
		"--gameDir " + minecraftPath + " " +
		"--assetIndex " + versionDetails.AssetIndex.ID + " " +
		"--assetsDir " + assetsPath() + " " +
		"--accessToken " + accessToken + " " +
		"--version " + versionDetails.ID))

	defer file.Close()
	return err
}
