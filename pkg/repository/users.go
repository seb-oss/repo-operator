package repository

import (
	"encoding/json"
	"math/rand"
	"time"
)

// User represents a user in artifactory
type User struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// UserAPIKey represents the JSON returned for a user's API Key in Artifactory
type UserAPIKey struct {
	APIKey string `json:"apiKey"`
}

// UserDetails represents the details of a user in artifactory
type UserDetails struct {
	Name                     string   `json:"name,omitempty"`
	Email                    string   `json:"email"`
	Password                 string   `json:"password"`
	Admin                    bool     `json:"admin,omitempty"`
	ProfileUpdatable         bool     `json:"profileUpdatable,omitempty"`
	DisableUIAccess          bool     `json:"disableUIAccess,omitempty"`
	InternalPasswordDisabled bool     `json:"internalPasswordDisabled,omitempty"`
	LastLoggedIn             string   `json:"lastLoggedIn,omitempty"`
	Realm                    string   `json:"realm,omitempty"`
	Groups                   []string `json:"groups,omitempty"`
}

// RepositoryUser struct
type RepositoryUser struct {
	Name                     string    `json:"name"`
	Email                    string    `json:"email"`
	Admin                    bool      `json:"admin"`
	ProfileUpdatable         bool      `json:"profileUpdatable"`
	InternalPasswordDisabled bool      `json:"internalPasswordDisabled"`
	Groups                   []string  `json:"groups"`
	LastLoggedIn             time.Time `json:"lastLoggedIn"`
	LastLoggedInMillis       int       `json:"lastLoggedInMillis"`
	Realm                    string    `json:"realm"`
	OfflineMode              bool      `json:"offlineMode"`
	DisableUIAccess          bool      `json:"disableUIAccess"`
}

// GetUser : Get user
func (R RTFactory) GetUser(c *Client, key string, q map[string]string) (RepositoryUser, int, string, error) {
	var res RepositoryUser
	userDetails, code, status, err := Get(c, "/api/security/users/"+key, q)
	if err != nil {
		return res, 500, statusInternalServerErrorState, err
	}
	err = json.Unmarshal(userDetails, &res)
	return res, code, status, err
}

// CreateUser creates a user with the specified details
func (R RTFactory) CreateUser(c *Client, key string, u UserDetails, q map[string]string) (int, string, error) {
	code := 0
	status := ""
	j, err := json.Marshal(u)
	if err != nil {
		return 500, statusInternalServerErrorState, err
	}
	_, code, status, err = Put(c, "/api/security/users/"+key, j, q)
	return code, status, err
}

// DeleteUser deletes a user
func (R RTFactory) DeleteUser(c *Client, key string) (int, string, error) {
	code := 0
	status := ""
	code, status, err := Delete(c, "/api/security/users/"+key)
	return code, status, err
}

// GenerateRandomPassword : Generates a Random password for internal user
func GenerateRandomPassword() string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	length := 8
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	for i := len(buf) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		buf[i], buf[j] = buf[j], buf[i]
	}
	str := string(buf)
	return str
}
