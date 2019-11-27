package repository

import (
	"encoding/json"
)

// PermissionTarget represents the json returned by Artifactory for a permission target
type PermissionTarget struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// PermissionTargetDetails represents the json returned by Artifactory for permission target details
type PermissionTargetDetails struct {
	Name            string     `json:"name,omitempty"`
	IncludesPattern string     `json:"includesPattern,omitempty"`
	ExcludesPattern string     `json:"excludesPattern,omitempty"`
	Repositories    []string   `json:"repositories,omitempty"`
	Principals      Principals `json:"principals,omitempty"`
}

// Principals represents the json response for principals in Artifactory
type Principals struct {
	Users  map[string][]string `json:"users"`
	Groups map[string][]string `json:"groups"`
}

// GetPermissionTargetDetails : get details about the permission target
func (R RTFactory) GetPermissionTargetDetails(c *Client, key string, q map[string]string) (PermissionTargetDetails, int, string, error) {
	var res PermissionTargetDetails
	permission, code, status, err := Get(c, "/api/security/permissions/"+key, q)
	if err != nil {
		return res, 500, statusInternalServerErrorState, err
	}
	err = json.Unmarshal(permission, &res)
	return res, code, status, err
}

// CreatePermissionTarget creates the named permission target
func (R RTFactory) CreatePermissionTarget(c *Client, key string, p PermissionTargetDetails, q map[string]string) (int, string, error) {
	j, err := json.Marshal(p)
	if err != nil {
		return 500, statusInternalServerErrorState, err
	}
	_, code, status, err := Put(c, "/api/security/permissions/"+key, j, q)
	return code, status, err
}

// DeletePermissionTarget : Delete permission target
func (R RTFactory) DeletePermissionTarget(c *Client, key string) (int, string, error) {
	var err error
	code := 0
	status := ""
	code, status, err = Delete(c, "/api/security/permissions/"+key)
	return code, status, err
}
