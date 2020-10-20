package model

type VersionManifest struct {
	Latest   Latest    `json:"latest"`
	Versions []Version `json:"versions"`
}

type Latest struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type Version struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
}

type VersionDetails struct {
	Arguments              map[string]interface{} `json:"arguments"`
	AssetIndex             AssetIndex             `json:"assetIndex"`
	Assets                 string                 `json:"assets"`
	ComplianceLevel        int                    `json:"complianceLevel"`
	Downloads              Downloads              `json:"downloads"`
	ID                     string                 `json:"id"`
	Libraries              []Library              `json:"libraries"`
	Logging                Logging                `json:"logging"`
	MainClass              string                 `json:"mainClass"`
	MinimumLauncherVersion int                    `json:"minimumLauncherVersion"`
	ReleaseTime            string                 `json:"releaseTime"`
	Time                   string                 `json:"time"`
	Type                   string                 `json:"type"`
}

type AssetIndex struct {
	ID        string `json:"id"`
	SHA1      string `json:"sha1"`
	Size      int    `json:"size"`
	TotalSize int    `json:"totalSize"`
	URL       string `json:"url"`
}

type Downloads struct {
	Client         DownloadItem `json:"client"`
	ClientMappings DownloadItem `json:"client_mappings"`
	Server         DownloadItem `json:"server"`
	ServerMappings DownloadItem `json:"serverMappings"`
}

type DownloadItem struct {
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

type Library struct {
	Name      string       `json:"name"`
	Downloads LibDownloads `json:"downloads"`
}

type LibDownloads struct {
	Artifact Artifact `json:"artifact"`
}

type Artifact struct {
	Path string `json:"path"`
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

type Logging struct {
	Client LoggingClient `json:"client"`
}

type LoggingClient struct {
	Argument string `json:"argument"`
	File     File   `json:"file"`
	Type     string `json:"type"`
}

type File struct {
	ID   string `json:"id"`
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}
