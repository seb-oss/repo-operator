package repository

import (
	"net/http"
	"testing"
)

func TestClient_CreateRepositories(t *testing.T) {

	client := &Client{
		Client:    nil,
		Config:    nil,
		Transport: nil,
		rt:        &mockArtifactoryClient{},
	}

	type fields struct {
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
	}
	type args struct {
		repoName   string
		repoType   string
		namespace  string
		statusCode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		code    int
		status  string
		wantErr bool
	}{
		{
			name:   "Test docker repo creation",
			fields: fields{},
			args: args{
				repoName:   "test-repo",
				repoType:   "docker",
				namespace:  "test-namespace",
				statusCode: 0,
			},
			code:    okStateCode,
			status:  statusOKState,
			wantErr: false,
		},
		{
			name:   "Test Maven repo creation",
			fields: fields{},
			args: args{
				repoName:   "test-repo",
				repoType:   "maven",
				namespace:  "test-namespace",
				statusCode: 0,
			},
			code:    okStateCode,
			status:  statusOKState,
			wantErr: false,
		},
		{
			name:   "Test Other repo creation",
			fields: fields{},
			args: args{
				repoName:   "test-repo",
				repoType:   "npm",
				namespace:  "test-namespace",
				statusCode: 0,
			},
			code:    okStateCode,
			status:  statusOKState,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := client.CreateRepositories(tt.args.repoName, tt.args.repoType, tt.args.namespace, tt.args.statusCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRepositories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.code {
				t.Errorf("CreateRepositories() got = %v, want %v", got, tt.code)
			}
			if got1 != tt.status {
				t.Errorf("CreateRepositories() got1 = %v, want %v", got1, tt.status)
			}
		})
	}
}

type mockArtifactoryClient struct{}

func (R mockArtifactoryClient) GetLocalRepo(c *Client, key string, q map[string]string) (RepositoryConfig, int, string, error) {
	gr := LocalRepoConfig{
		GenericRepoConfig: GenericRepoConfig{
			Key:    "test-repository",
			RClass: "maven",
		},
	}
	return gr, okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) GetVirtualRepo(c *Client, key string, q map[string]string) (RepositoryConfig, int, string, error) {
	gr := VirtualRepoConfig{
		GenericRepoConfig: GenericRepoConfig{
			Key:    "test-repository",
			RClass: "maven",
		},
	}
	return gr, okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) GetRemoteRepos(c *Client, packageType string) ([]RemoteRepo, int, string, error) {
	rr := []RemoteRepo{
		{
			Key:   "remote-repo1",
			Rtype: "maven",
		},
		{
			Key:   "remote-repo2",
			Rtype: "maven",
		},
	}
	return rr, okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) CreateRepo(c *Client, key string, r RepositoryConfig, q map[string]string) (int, string, error) {
	return okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) DeleteRepo(c *Client, key string) (int, string, error) {
	return okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) GetUser(c *Client, key string, q map[string]string) (RepositoryUser, int, string, error) {
	ru := RepositoryUser{
		Name: "repo-test-user",
	}
	return ru, okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) CreateUser(c *Client, key string, u UserDetails, q map[string]string) (int, string, error) {
	return okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) DeleteUser(c *Client, key string) (int, string, error) {
	return okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) GetPermissionTargetDetails(c *Client, key string, q map[string]string) (PermissionTargetDetails, int, string, error) {
	userList := []string{"test-user"}
	u := map[string][]string{}
	for _, each := range userList {
		u[each] = []string{"r", "d", "w", "n", "m"}
	}
	pr := PermissionTargetDetails{
		Name:            "test-permission",
		IncludesPattern: "",
		ExcludesPattern: "",
		Repositories:    nil,
		Principals: Principals{
			Users: u,
		},
	}
	return pr, okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) CreatePermissionTarget(c *Client, key string, p PermissionTargetDetails, q map[string]string) (int, string, error) {
	return okStateCode, statusOKState, nil
}

func (R mockArtifactoryClient) DeletePermissionTarget(c *Client, key string) (int, string, error) {
	return okStateCode, statusOKState, nil
}

func TestClient_CreateRepositoryUser(t *testing.T) {
	client := &Client{
		Client:    nil,
		Config:    nil,
		Transport: nil,
		rt:        &mockArtifactoryClient{},
	}
	type fields struct {
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
		rt        rtInterface
	}
	type args struct {
		reqName string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		password string
		code     int
		status   string
		wantErr  bool
	}{
		{
			name: "Test Internal user creation",
			args: args{
				reqName: "test-repo",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _, _, err := client.CreateRepositoryUser(tt.args.reqName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRepositoryUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Check if random password is generated
			if p == "" {
				t.Errorf("Random password is not generated properly")
				return
			}
		})
	}
}

func TestClient_CleanupRepository(t *testing.T) {
	client := &Client{
		Client:    nil,
		Config:    nil,
		Transport: nil,
		rt:        &mockArtifactoryClient{},
	}
	type fields struct {
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
		rt        rtInterface
	}
	type args struct {
		reqName   string
		repoType  string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Test Maven Cleanup repository",
			fields: fields{},
			args: args{
				reqName:   "test-repo",
				repoType:  "maven",
				namespace: "test-namespace",
			},
			wantErr: false,
		},
		{
			name:   "Test Docker Cleanup repository",
			fields: fields{},
			args: args{
				reqName:   "test-repo",
				repoType:  "docker",
				namespace: "test-namespace",
			},
			wantErr: false,
		},
		{
			name:   "Test other Cleanup repository",
			fields: fields{},
			args: args{
				reqName:   "test-repo",
				repoType:  "npm",
				namespace: "test-namespace",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.CleanupRepository(tt.args.reqName, tt.args.repoType, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("CleanupRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreatePermission(t *testing.T) {
	client := &Client{
		Client:    nil,
		Config:    nil,
		Transport: nil,
		rt:        &mockArtifactoryClient{},
	}
	type fields struct {
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
		rt        rtInterface
	}
	type args struct {
		reqName      string
		repoType     string
		namespace    string
		users        []string
		repositories []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Test create permission with no user",
			fields: fields{},
			args: args{
				reqName:      "test-repo",
				repoType:     "maven",
				namespace:    "test-namespace",
				users:        nil,
				repositories: nil,
			},
			wantErr: false,
		},
		{
			name:   "Test create permission with user",
			fields: fields{},
			args: args{
				reqName:   "test-repo",
				repoType:  "maven",
				namespace: "test-namespace",
				users: []string{
					"test-user",
				},
				repositories: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := client.CreatePermission(tt.args.reqName, tt.args.repoType, tt.args.namespace, tt.args.users, tt.args.repositories); (err != nil) != tt.wantErr {
				t.Errorf("CreatePermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
