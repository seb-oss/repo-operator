package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRTFactory_GetLocalRepo(t *testing.T) {
	r := LocalRepoConfig{
		GenericRepoConfig: GenericRepoConfig{
			Key: "test-repo",
		},
	}
	responseBody, _ := json.Marshal(r)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, string(responseBody))
		if err != nil {
			t.Error("Failed to setup server")
		}
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)
	type fields struct {
		rtInterface rtInterface
	}
	type args struct {
		c   *Client
		key string
		q   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    RepositoryConfig
		want1   int
		want2   string
		wantErr bool
	}{
		{
			name:   "Test Get local repo",
			fields: fields{},
			args: args{
				key: "test-repo",
				q:   nil,
			},
			want:    nil,
			want1:   200,
			want2:   statusOKState,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			R := RTFactory{
				rtInterface: tt.fields.rtInterface,
			}
			_, _, _, err := R.GetLocalRepo(&client, tt.args.key, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRTFactory_GetVirtualRepo(t *testing.T) {
	r := VirtualRepoConfig{
		GenericRepoConfig: GenericRepoConfig{
			Key: "test-repo",
		},
	}
	responseBody, _ := json.Marshal(r)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, string(responseBody))
		if err != nil {
			t.Error("Failed to setup server")
		}
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)
	type fields struct {
		rtInterface rtInterface
	}
	type args struct {
		c   *Client
		key string
		q   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    RepositoryConfig
		want1   int
		want2   string
		wantErr bool
	}{
		{
			name:   "Test Get Remote repo",
			fields: fields{},
			args: args{
				key: "test-repo",
				q:   nil,
			},
			want:    nil,
			want1:   200,
			want2:   statusOKState,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			R := RTFactory{
				rtInterface: tt.fields.rtInterface,
			}
			_, _, _, err := R.GetVirtualRepo(&client, tt.args.key, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRTFactory_GetRemoteRepos(t *testing.T) {
	r := []RemoteRepo{
		{
			Key:   "remote-repo1",
			Rtype: "maven",
		},
	}
	responseBody, _ := json.Marshal(r)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, string(responseBody))
		if err != nil {
			t.Error("Failed to setup server")
		}
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)

	type fields struct {
		rtInterface rtInterface
	}
	type args struct {
		packageType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []RemoteRepo
		wantErr bool
	}{
		{
			name:   "Get Remote repos",
			fields: fields{},
			args: args{
				packageType: "maven",
			},
			want: []RemoteRepo{
				{
					Key:   "remote-repo1",
					Rtype: "maven",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			R := RTFactory{
				rtInterface: tt.fields.rtInterface,
			}
			got, _, _, err := R.GetRemoteRepos(&client, tt.args.packageType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRemoteRepos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got[0].Key != "remote-repo1" {
				t.Errorf("Incorrect remote repo")
				return
			}
		})
	}
}

func TestRTFactory_CreateRepo(t *testing.T) {
	var buf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		req, _ := ioutil.ReadAll(r.Body)
		buf.Write(req)
		_, err := fmt.Fprintf(w, "")
		if err != nil {
			t.Error("Failed to setup server")
		}
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)
	type fields struct {
		rtInterface rtInterface
	}
	type args struct {
		c   *Client
		key string
		r   RepositoryConfig
		q   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		want1   string
		wantErr bool
	}{
		{
			name:   "Test local repository creation",
			fields: fields{},
			args: args{
				key: "test-repo",
				r: LocalRepoConfig{
					GenericRepoConfig: GenericRepoConfig{
						Key: "test-repo",
					},
				},
				q: nil,
			},
			want:    0,
			want1:   "",
			wantErr: false,
		},
		{
			name:   "Test virtual repository creation",
			fields: fields{},
			args: args{
				key: "test-repo",
				r: VirtualRepoConfig{
					GenericRepoConfig: GenericRepoConfig{
						Key: "test-repo",
					},
				},
				q: nil,
			},
			want:    0,
			want1:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			R := RTFactory{
				rtInterface: tt.fields.rtInterface,
			}
			_, _, err := R.CreateRepo(&client, tt.args.key, tt.args.r, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRTFactory_DeleteRepo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, "")
		if err != nil {
			t.Error("Failed to setup server")
		}
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
		Transport: transport,
	}

	client := NewClient(conf)
	type fields struct {
		rtInterface rtInterface
	}
	type args struct {
		c   *Client
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		want1   string
		wantErr bool
	}{
		{
			name:   "Test repository deletion",
			fields: fields{},
			args: args{
				key: "test-repo",
			},
			want:    0,
			want1:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			R := RTFactory{
				rtInterface: tt.fields.rtInterface,
			}
			_, _, err := R.DeleteRepo(&client, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
