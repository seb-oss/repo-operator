package repository

import (
	"encoding/json"
)

// Repo represents the json response from Artifactory describing a repository
type Repo struct {
	Key         string `json:"key"`
	Rtype       string `json:"type"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// RemoteRepo : Repo represents the json response from Artifactory describing a repository
type RemoteRepo struct {
	Key         string `json:"key"`
	Rtype       string `json:"type"`
	URL         string `json:"url,omitempty"`
	PackageType string `json:"packagetype,omitempty"`
}

// RepositoryConfig represents a repo config
type RepositoryConfig interface {
	MimeType() string
}

// GenericRepoConfig represents the common json of a repo response from artifactory
type GenericRepoConfig struct {
	Key                          string   `json:"key,omitempty"`
	RClass                       string   `json:"rclass"`
	PackageType                  string   `json:"packageType,omitempty"`
	Description                  string   `json:"description,omitempty"`
	Notes                        string   `json:"notes,omitempty"`
	IncludesPattern              string   `json:"includesPattern,omitempty"`
	ExcludesPattern              string   `json:"excludesPattern,omitempty"`
	LayoutRef                    string   `json:"repoLayoutRef,omitempty"`
	HandleReleases               *bool    `json:"handleReleases,omitempty"`
	HandleSnapshots              *bool    `json:"handleSnapshots,omitempty"`
	MaxUniqueSnapshots           int      `json:"maxUniqueSnapshots,omitempty"`
	SuppressPomConsistencyChecks bool     `json:"suppressPomConsistencyChecks,omitempty"`
	BlackedOut                   bool     `json:"blackedOut,omitempty"`
	PropertySets                 []string `json:"propertySets,omitempty"`
}

// MimeType returns the MimeType of a GenericRepoConfig
func (r GenericRepoConfig) MimeType() string {
	return ""
}

// LocalRepoConfig represents a local repo type in artifactory
type LocalRepoConfig struct {
	GenericRepoConfig

	DebianTrivialLayout     bool   `json:"debianTrivialLayout,omitempty"`
	ChecksumPolicyType      string `json:"checksumPolicyType,omitempty"`
	MaxUniqueTags           int    `json:"maxUniqueTags,omitempty"`
	SnapshotVersionBehavior string `json:"snapshotVersionBehavior,omitempty"`
	ArchiveBrowsingEnabled  bool   `json:"archiveBrowsingEnabled,omitempty"`
	CalculateYumMetadata    bool   `json:"calculateYumMetadata,omitempty"`
	YumRootDepth            int    `json:"yumRootDepth,omitempty"`
	DockerAPIVersion        string `json:"dockerApiVersion,omitempty"`
	EnableFileListsIndexing bool   `json:"enableFileListsIndexing,omitempty"`
	XrayIndex               bool   `json:"xrayIndex,omitempty"`
}

// MimeType returns the MimeType for a local repo in artifactory
func (r LocalRepoConfig) MimeType() string {
	return LocalRepoMimeType
}

// RemoteRepoConfig represents a remote repo in artifactory
type RemoteRepoConfig struct {
	GenericRepoConfig

	URL                               string `json:"url"`
	Username                          string `json:"username,omitempty"`
	Password                          string `json:"password,omitempty"`
	Proxy                             string `json:"proxy,omitempty"`
	RemoteRepoChecksumPolicyType      string `json:"remoteRepoChecksumPolicyType,omitempty"`
	HardFail                          bool   `json:"hardFail,omitempty"`
	Offline                           bool   `json:"offline,omitempty"`
	StoreArtifactsLocally             bool   `json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis               int    `json:"socketTimeoutMillis,omitempty"`
	LocalAddress                      string `json:"localAddress,omitempty"`
	RetrivialCachePeriodSecs          int    `json:"retrievalCachePeriodSecs,omitempty"`
	FailedRetrievalCachePeriodSecs    int    `json:"failedRetrievalCachePeriodSecs,omitempty"`
	MissedRetrievalCachePeriodSecs    int    `json:"missedRetrievalCachePeriodSecs,omitempty"`
	UnusedArtifactsCleanupEnabled     bool   `json:"unusedArtifactCleanupEnabled,omitempty"`
	UnusedArtifactsCleanupPeriodHours int    `json:"unusedArtifactCleanupPeriodHours,omitempty"`
	FetchJarsEagerly                  bool   `json:"fetchJarsEagerly,omitempty"`
	FetchSourcesEagerly               bool   `json:"fetchSourcesEagerly,omitempty"`
	ShareConfiguration                bool   `json:"shareConfiguration,omitempty"`
	SynchronizeProperties             bool   `json:"synchronizeProperties,omitempty"`
	BlockMismatchingMimeTypes         bool   `json:"blockMismatchingMimeTypes,omitempty"`
	AllowAnyHostAuth                  bool   `json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            bool   `json:"enableCookieManagement,omitempty"`
	BowerRegistryURL                  string `json:"bowerRegistryUrl,omitempty"`
	VcsType                           string `json:"vcsType,omitempty"`
	VcsGitProvider                    string `json:"vcsGitProvider,omitempty"`
	VcsGitDownloader                  string `json:"vcsGitDownloader,omitempty"`
	ClientTLSCertificate              string `json:"clientTlsCertificate,omitempty"`
}

// MimeType returns the mimetype of a remote repo
func (r RemoteRepoConfig) MimeType() string {
	return RemoteRepoMimeType
}

// VirtualRepoConfig represents a virtual repo in artifactory
type VirtualRepoConfig struct {
	GenericRepoConfig

	Repositories                                  []string `json:"repositories"`
	DebianTrivialLayout                           bool     `json:"debianTrivialLayout,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts bool     `json:"artifactoryRequestsCanRetrieveRemoteArtifacts,omitempty"`
	KeyPair                                       string   `json:"keyPair,omitempty"`
	PomRepositoryReferencesCleanupPolicy          string   `json:"pomRepositoryReferencesCleanupPolicy,omitempty"`
	DefaultDeploymentRepo                         string   `json:"defaultDeploymentRepo,omitempty"`
}

// MimeType returns the mimetype for a virtual repo in artifactory
func (r VirtualRepoConfig) MimeType() string {
	return VirtualRepoMimeType
}

// GetRemoteRepos returns all repos of the provided type
func (R RTFactory) GetRemoteRepos(c *Client, packageType string) ([]RemoteRepo, int, string, error) {
	o := make(map[string]string)
	cd := 0
	o["type"] = "remote"
	o["packageType"] = packageType
	var dat []RemoteRepo
	d, cd, status, err := Get(c, "/api/repositories", o)
	if err != nil {
		return dat, 500, statusInternalServerErrorState, err
	}
	err = json.Unmarshal(d, &dat)
	if err != nil {
		return dat, 500, statusInternalServerErrorState, err
	}
	return dat, cd, status, err
}

// GetLocalRepo returns the named local repo
func (R RTFactory) GetLocalRepo(c *Client, key string, q map[string]string) (RepositoryConfig, int, string, error) {
	var dat GenericRepoConfig
	d, code, status, err := Get(c, "/api/repositories/"+key, q)
	if err != nil {
		return dat, 500, statusInternalServerErrorState, err
	}
	err = json.Unmarshal(d, &dat)
	if err != nil {
		return dat, 500, statusInternalServerErrorState, err
	}
	var cdat LocalRepoConfig
	_ = json.Unmarshal(d, &cdat)
	return cdat, code, status, nil
}

// GetLocalRepo returns the named local repo
func (R RTFactory) GetVirtualRepo(c *Client, key string, q map[string]string) (RepositoryConfig, int, string, error) {
	var dat GenericRepoConfig
	d, code, status, err := Get(c, "/api/repositories/"+key, q)
	if err != nil {
		return dat, 500, statusInternalServerErrorState, err
	}
	err = json.Unmarshal(d, &dat)
	if err != nil {
		return dat, 500, statusInternalServerErrorState, err
	}
	var cdat VirtualRepoConfig
	_ = json.Unmarshal(d, &cdat)
	return cdat, code, status, nil
}

// CreateRepo creates the named repo
func (R RTFactory) CreateRepo(c *Client, key string, r RepositoryConfig, q map[string]string) (int, string, error) {
	code := 0
	status := ""
	j, err := json.Marshal(r)
	if err != nil {
		return 500, statusInternalServerErrorState, err
	}
	_, code, status, err = Put(c, "/api/repositories/"+key, j, q)
	return code, status, err
}

// DeleteRepo creates the named repo
func (R RTFactory) DeleteRepo(c *Client, key string) (int, string, error) {
	var err error
	code := 0
	status := ""
	code, status, err = Delete(c, "/api/repositories/"+key)
	return code, status, err
}
