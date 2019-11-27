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

func TestClient_CreatePermissionTarget(t *testing.T) {
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
	permTarget := PermissionTargetDetails{
		Name:            "test-permission",
		IncludesPattern: "**",
		ExcludesPattern: "",
		Repositories:    []string{"docker-local-v2", "libs-release-local", "plugins-release-local"},
		Principals: Principals{
			Users:  map[string][]string{"admin": []string{"r", "d", "w", "n", "m"}},
			Groups: map[string][]string{"java-committers": []string{"r", "d", "w", "n"}},
		},
	}
	type fields struct {
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
	}
	type args struct {
		key string
		p   PermissionTargetDetails
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
			name: "Create Permission object",
			args: args{
				key: "test-permission",
				p:   permTarget,
				q:   make(map[string]string),
			},
			want:    200,
			want1:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, status, err := client.rt.CreatePermissionTarget(&client, tt.args.key, tt.args.p, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePermissionTarget() returned error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if code != 200 {
				t.Errorf("CreatePermissionTarget() create permission failed with code = %v, want %v", code, tt.want)
			}
			if status == statusInternalServerErrorState {
				t.Errorf("CreatePermissionTarget() status = %v, want %v", code, tt.want)
			}
		})
	}
}

func TestClient_DeletePermissionTarget(t *testing.T) {
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
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Delete Permission object",
			args: args{
				key: "test-permission",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := client.rt.DeletePermissionTarget(&client, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePermissionTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClient_GetPermissionTargetDetails(t *testing.T) {
	res := PermissionTargetDetails{
		Name:            "test-permission",
		IncludesPattern: "**",
		ExcludesPattern: "",
		Repositories:    []string{"docker-local-v2", "libs-release-local", "plugins-release-local"},
		Principals: Principals{
			Users:  map[string][]string{"admin": []string{"r", "d", "w", "n", "m"}},
			Groups: map[string][]string{"java-committer": []string{"r", "d", "w", "n"}},
		},
	}
	responseBody, _ := json.Marshal(res)
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
		Client    *http.Client
		Config    *ClientConfig
		Transport *http.Transport
	}
	type args struct {
		key string
		q   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PermissionTargetDetails
		want1   int
		want2   string
		wantErr bool
	}{
		{
			name: "Get permission object",
			args: args{
				key: "test-permission",
				q:   make(map[string]string),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, _, _, err := client.rt.GetPermissionTargetDetails(&client, tt.args.key, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPermissionTargetDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pt.Name != res.Name {
				t.Errorf("GetPermissionTargetDetails() permission details name should be error = %v, wantErr %v", err, res.Name)
			}
		})
	}
}
