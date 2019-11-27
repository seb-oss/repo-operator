package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_TestCreateUser(t *testing.T) {
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

	user := UserDetails{
		Name:                     "admin",
		Email:                    "test@test.com",
		Password:                 "somepass",
		Admin:                    true,
		ProfileUpdatable:         true,
		DisableUIAccess:          false,
		InternalPasswordDisabled: false,
		Groups:                   []string{"administrators"},
		LastLoggedIn:             "2015-08-11T14:04:11.472Z",
		Realm:                    "internal",
	}

	expectedJSON, _ := json.Marshal(user)
	_, _, err := client.rt.CreateUser(&client, "admin", user, make(map[string]string))
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, string(expectedJSON), buf.String(), "should send user json")
}

func TestClient_TestCreateUserFailure(t *testing.T) {
	conf := &ClientConfig{
		BaseURL:   "http://127.0.0.1:8080/",
		Username:  "username",
		Password:  "password",
		VerifySSL: false,
	}

	client := NewClient(conf)
	var details = UserDetails{}
	_, _, err := client.rt.CreateUser(&client, "testuser", details, make(map[string]string))
	assert.Error(t, err, "should return an error")
}

func TestClient_TestDeleteUser(t *testing.T) {
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
	_, _, err := client.rt.DeleteUser(&client, "testuser")
	assert.NoError(t, err, "should not return an error")
}

func TestRTFactory_GetUser(t *testing.T) {
	res := RepositoryUser{
		Name: "test-user",
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
		want    RepositoryUser
		wantErr bool
	}{
		{
			name:   "Test get user",
			fields: fields{},
			args: args{
				key: "test-user",
				q:   nil,
			},
			want: RepositoryUser{
				Name: "test-user",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, _, err := client.rt.GetUser(&client, tt.args.key, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.want.Name {
				t.Error("Returned user is not the same")
			}
		})
	}
}
