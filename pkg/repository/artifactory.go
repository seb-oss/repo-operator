package repository

import (
	"github.com/go-logr/logr"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sort"
	"strconv"
	"strings"
)

var log = logf.Log.WithName("controller_repository")
var Code int
var Status string

const (
	ins                             = "instance.Namespace"
	rname                           = "req.Name"
	artifactoryClassLocal           = "local"
	artifactoryClassVirtual         = "virtual"
	dockerRepoType                  = "docker"
	mavenRepoType                   = "maven"
	okStateCode                     = 200
	conflictStateCode               = 409
	statusOKState                   = "ok"
	conflictState                   = "Conflict"
	suffixPackageClassLocal         = "-local"
	suffixArtifactoryRepoUser       = "-repo-user"
	suffixArtifactoryRepoPermission = "-repo-permission"
	snapshotSuffix                  = "-snapshot"
	releaseSuffix                   = "-release"
	statusInternalServerErrorState  = "Internal Server Error"
	errorFailedToDeleteRepo         = "failed to delete repo "
	namespaceInhouseLibraries       = " namespace in-house libraries"
	localRepositoryFor              = "Local repository for "
)

func init() {
	Code = okStateCode
	Status = statusOKState
}

type RTFactory struct {
	rtInterface
}

type rtInterface interface {
	GetLocalRepo(c *Client, key string, q map[string]string) (RepositoryConfig, int, string, error)
	GetVirtualRepo(c *Client, key string, q map[string]string) (RepositoryConfig, int, string, error)
	GetRemoteRepos(c *Client, packageType string) ([]RemoteRepo, int, string, error)
	CreateRepo(c *Client, key string, r RepositoryConfig, q map[string]string) (int, string, error)
	DeleteRepo(c *Client, key string) (int, string, error)
	GetUser(c *Client, key string, q map[string]string) (RepositoryUser, int, string, error)
	CreateUser(c *Client, key string, u UserDetails, q map[string]string) (int, string, error)
	DeleteUser(c *Client, key string) (int, string, error)
	GetPermissionTargetDetails(c *Client, key string, q map[string]string) (PermissionTargetDetails, int, string, error)
	CreatePermissionTarget(c *Client, key string, p PermissionTargetDetails, q map[string]string) (int, string, error)
	DeletePermissionTarget(c *Client, key string) (int, string, error)
}

// CreateRepositories : Function creates all the required repositories
func (c *Client) CreateRepositories(repoName string, repoType string, namespace string, statusCode int) (int, string, error) {
	var repoLocal RepositoryConfig
	var repoVirtual RepositoryConfig
	var remoteRepos []RemoteRepo

	localRepoExist, codeLocal, statusLocal, err := c.createLocalRepository(repoLocal, repoName, repoType, namespace)
	if err != nil {
		return codeLocal, statusLocal, err
	}

	virtualRepoExist, codeVirtual, statusVirtual, err := c.createVirtualRepository(repoVirtual, repoName, remoteRepos, repoType, namespace)
	if err != nil {
		return codeVirtual, statusVirtual, err
	}

	// Check if repo already exist with the name and use don't accidentally modify some other repo
	if localRepoExist && virtualRepoExist && statusCode != 200 {
		return conflictStateCode, conflictState, nil
	}
	return okStateCode, statusOKState, nil
}

// CreateRepositories : Function creates virtual repositories
func (c *Client) createVirtualRepository(repoVirtual RepositoryConfig, repoName string, remoteRepos []RemoteRepo, repoType string, namespace string) (bool, int, string, error) {
	// Check if Virtual Repository already exists in Artifactory
	virtualRepoExist := false
	reqLogger := log.WithValues(ins, namespace, rname, repoName)
	repoVirtual, Code, Status, err := c.rt.GetVirtualRepo(c, repoName, make(map[string]string))
	if err != nil {
		return false, Code, Status, err
	}
	if repoVirtual.(VirtualRepoConfig).Key == repoName {
		// Repository already exists - don't requeue
		reqLogger.Info("Skip reconcile: If repo already exists " + repoVirtual.(VirtualRepoConfig).Key)
		virtualRepoExist = true
	}
	if !virtualRepoExist {
		// Get all the remote repositories for particular type.
		remoteRepos, Code, Status, err = c.rt.GetRemoteRepos(c, repoType)
		if err != nil {
			return false, Code, Status, err
		}
		repositories := []string{}
		for _, repos := range remoteRepos {
			repositories = append(repositories, repos.Key)
		}

		reqLogger.Info("Creating virtual repository...." + repoName)
		// Create repository if it doesn't exist
		Code, Status, err = c.rt.CreateRepo(c, repoName, getVirtualRepoConfig(repositories, repoName, repoType, namespace, artifactoryClassVirtual), make(map[string]string))
		if err != nil {
			return false, Code, Status, err
		}
	}
	return virtualRepoExist, 0, "", nil
}

// CreateRepositories : Function creates Local repositories
func (c *Client) createLocalRepository(repoLocal RepositoryConfig, repoName string, repoType string, namespace string) (bool, int, string, error) {
	// Check if local Repository already exists in Artifactory
	localRepoExist := false
	reqLogger := log.WithValues(ins, namespace, rname, repoName)
	repoLocal, Code, Status, err := c.rt.GetLocalRepo(c, repoName+suffixPackageClassLocal, make(map[string]string))
	if err != nil {
		return false, Code, Status, nil
	}
	if repoLocal.(LocalRepoConfig).Key == repoName+suffixPackageClassLocal {
		// Repository already exists - don't requeue
		reqLogger.Info("Skip reconcile: If repo already exists " + repoLocal.(LocalRepoConfig).Key)
		localRepoExist = true
	}
	if !localRepoExist {
		reqLogger.Info("Creating local repository...." + repoName + suffixPackageClassLocal)
		// Create repository if it doesn't exist
		Code, Status, err = c.rt.CreateRepo(c, repoName+suffixPackageClassLocal, getLocalRepoConfig(repoName, repoType, namespace, artifactoryClassLocal), make(map[string]string))
		if err != nil {
			return false, Code, Status, err
		}
	}
	return localRepoExist, 0, "", nil
}

// CreateRepositoryUser : Create repository user
func (c *Client) CreateRepositoryUser(reqName string) (string, int, string, error) {
	// Generate random password
	rp := GenerateRandomPassword()
	userDetails := UserDetails{
		Name:                     reqName + suffixArtifactoryRepoUser,
		Email:                    reqName + suffixArtifactoryRepoUser + "@internal.com",
		Password:                 rp,
		DisableUIAccess:          true,
		ProfileUpdatable:         false,
		InternalPasswordDisabled: false,
		Realm:                    "Internal",
	}
	cd, s, err := c.rt.CreateUser(c, reqName+suffixArtifactoryRepoUser, userDetails, make(map[string]string))
	return rp, cd, s, err
}

// CreatePermission : Create permission object
func (c *Client) CreatePermission(reqName string, repoType string, namespace string, users []string, repositories []string) error {
	reqLogger := log.WithValues(ins, namespace, rname, reqName)
	// Create Permission target for User
	reqLogger.Info("Create Permission target - "+reqName+"-"+repoType+suffixArtifactoryRepoPermission, "Namespace", namespace, "Name", reqName)
	// create user list to be added for access

	// Check if any user is Admin or remove user if not found
	userList, err, empty := c.filerUserList(users, reqLogger)
	if empty {
		return err
	}

	u := map[string][]string{}
	for _, each := range userList {
		u[each] = []string{"r", "d", "w", "n", "m"}
	}
	pt := PermissionTargetDetails{
		Name:            reqName + "-" + repoType + suffixArtifactoryRepoPermission,
		IncludesPattern: "**",
		ExcludesPattern: "",
		Repositories:    repositories,
		Principals: Principals{
			Users: u,
		},
	}
	// Check if permission object already exists in Artifactory
	ptd, _, _, err := c.rt.GetPermissionTargetDetails(c, reqName+"-"+repoType+suffixArtifactoryRepoPermission, make(map[string]string))
	if err != nil {
		reqLogger.Info("Permission target does not exist - it will be created")
		_, _, err := c.rt.CreatePermissionTarget(c, reqName+"-"+repoType+suffixArtifactoryRepoPermission, pt, make(map[string]string))
		if err != nil {
			reqLogger.Error(err, "failed to create permission target")
			return err
		}
	} else {
		reqLogger.Info("Permission target already exist check if there is any change in user...")
		existingUserList := []string{}
		for user := range ptd.Principals.Users {
			existingUserList = append(existingUserList, user)
		}
		//sort Maps before compare
		sort.Strings(existingUserList)
		sort.Strings(userList)
		if !reflect.DeepEqual(existingUserList, userList) {
			reqLogger.Info("Changes in the user list detected - update permission target")
			// Create or replace the permission target; this  should even work for creation and deletion of users
			_, _, err := c.rt.CreatePermissionTarget(c, reqName+"-"+repoType+suffixArtifactoryRepoPermission, pt, make(map[string]string))
			if err != nil {
				reqLogger.Error(err, "failed to update permission target")
				return err
			}
		} else {
			reqLogger.Info("No changes in user list - skip update")
		}
	}
	return nil
}

// Filter
func (c *Client) filerUserList(users []string, reqLogger logr.Logger) ([]string, error, bool) {
	userList := []string{}
	for _, user := range users {
		userDetails, _, _, err := c.rt.GetUser(c, user, make(map[string]string))
		if err != nil {
			reqLogger.Error(err, "failed to get user - do not add to list")
		}
		if !userDetails.Admin && err == nil {
			userList = append(userList, user)
		}
	}
	if len(userList) == 0 {
		reqLogger.Info("No users to add - not creating permission object")
		return nil, nil, true
	}
	return userList, nil, false
}

// CleanupRepository : It clean-up everything related to repositories
func (c *Client) CleanupRepository(reqName string, repoType string, namespace string) error {
	reqLogger := log.WithValues(ins, namespace, rname, reqName)
	switch repoType {
	case mavenRepoType:
		c.cleanUpMavenRepository(reqName, repoType, reqLogger)

	case dockerRepoType:
		c.cleanUpDockerRepository(reqName, repoType, reqLogger)

	default:
		c.cleanUpOtherRepository(reqName, repoType, reqLogger)
	}
	// Clean Permission Target
	_, _, err := c.rt.DeletePermissionTarget(c, reqName+"-"+repoType+suffixArtifactoryRepoPermission)
	if err != nil {
		reqLogger.Error(err, "failed to delete permission "+reqName+"-"+repoType+suffixArtifactoryRepoPermission)
		//return err
	}
	return nil
}

// Cleanup Maven repository
func (c *Client) cleanUpMavenRepository(reqName string, repoType string, reqLogger logr.Logger) {
	// cleanup local snapshot repos
	_, _, err := c.rt.DeleteRepo(c, reqName+"-"+repoType+snapshotSuffix+suffixPackageClassLocal)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType+snapshotSuffix+suffixPackageClassLocal)
		//return err
	}
	// cleanup virtual snapshot repos
	_, _, err = c.rt.DeleteRepo(c, reqName+"-"+repoType+snapshotSuffix)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType+snapshotSuffix)
		//return err
	}
	// cleanup local release repos
	_, _, err = c.rt.DeleteRepo(c, reqName+"-"+repoType+releaseSuffix+suffixPackageClassLocal)
	if err != nil {
		reqLogger.Error(err, "failed to delete repo"+reqName+"-"+repoType+releaseSuffix+suffixPackageClassLocal)
		//return err
	}
	// cleanup virtual release repos
	_, _, err = c.rt.DeleteRepo(c, reqName+"-"+repoType+releaseSuffix)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType+releaseSuffix)
		//return err
	}
}

// Cleanup Docker repository
func (c *Client) cleanUpDockerRepository(reqName string, repoType string, reqLogger logr.Logger) {
	// cleanup local repos
	_, _, err := c.rt.DeleteRepo(c, reqName+"-"+repoType+suffixPackageClassLocal)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType+suffixPackageClassLocal)
		//return err
	}
	// cleanup virtual repos
	_, _, err = c.rt.DeleteRepo(c, reqName+"-"+repoType)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType)
		//return err
	}
	// Clean User
	_, _, err = c.rt.DeleteUser(c, reqName+suffixArtifactoryRepoUser)
	if err != nil {
		reqLogger.Error(err, "failed to delete user")
		//return err
	}
}

// CLeanup other repository
func (c *Client) cleanUpOtherRepository(reqName string, repoType string, reqLogger logr.Logger) {
	// cleanup local repos
	_, _, err := c.rt.DeleteRepo(c, reqName+"-"+repoType+suffixPackageClassLocal)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType+suffixPackageClassLocal)
		//return err
	}
	// cleanup virtual repos
	_, _, err = c.rt.DeleteRepo(c, reqName+"-"+repoType)
	if err != nil {
		reqLogger.Error(err, errorFailedToDeleteRepo+reqName+"-"+repoType)
		//return err
	}
}

// Function to generate configuration for Local repositories.
func getLocalRepoConfig(repoName string, repoType string, namespace string, packageClass string) LocalRepoConfig {

	switch repoType {
	case mavenRepoType:
		snapshot, _ := strconv.ParseBool("true")
		release, _ := strconv.ParseBool("true")
		if strings.Contains(repoName, snapshotSuffix) {
			release, _ = strconv.ParseBool("false")
		} else if strings.Contains(repoName, releaseSuffix) {
			snapshot, _ = strconv.ParseBool("false")
		}

		rc := LocalRepoConfig{
			GenericRepoConfig: GenericRepoConfig{
				Key:             repoName + suffixPackageClassLocal,
				RClass:          packageClass,
				PackageType:     repoType,
				Description:     localRepositoryFor + namespace + namespaceInhouseLibraries,
				LayoutRef:       repoType + "-2-default",
				HandleSnapshots: &snapshot,
				HandleReleases:  &release,
			},
			XrayIndex: true,
		}
		return rc
	case dockerRepoType:
		rc := LocalRepoConfig{
			GenericRepoConfig: GenericRepoConfig{
				Key:         repoName + suffixPackageClassLocal,
				RClass:      packageClass,
				PackageType: repoType,
				Description: localRepositoryFor + namespace + namespaceInhouseLibraries,
				LayoutRef:   "simple-default",
			},
			XrayIndex: true,
		}
		return rc

	default:
		rc := LocalRepoConfig{
			GenericRepoConfig: GenericRepoConfig{
				Key:         repoName + suffixPackageClassLocal,
				RClass:      packageClass,
				PackageType: repoType,
				Description: localRepositoryFor + namespace + namespaceInhouseLibraries,
				LayoutRef:   repoType + "-default",
			},
			XrayIndex: true,
		}
		return rc
	}
}

// Function to generate configuration for Virtual repositories.
func getVirtualRepoConfig(remoteRepositories []string, repoName string, repoType string, namespace string, packageClass string) VirtualRepoConfig {
	repos := append(remoteRepositories, repoName+suffixPackageClassLocal)
	rc := VirtualRepoConfig{
		GenericRepoConfig: GenericRepoConfig{
			Key:          repoName,
			RClass:       packageClass,
			PackageType:  repoType,
			LayoutRef:    "simple-default",
			PropertySets: []string{"artifactory"},
			Description:  "virtual repository for " + namespace + " namespace and required remote libraries",
		},
		Repositories:          repos,
		DefaultDeploymentRepo: repoName + suffixPackageClassLocal,
	}
	return rc
}
