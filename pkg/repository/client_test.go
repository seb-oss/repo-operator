package repository

import (
	"net/http"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		config *ClientConfig
	}
	tests := []struct {
		name string
		args args
		want Client
	}{
		{
			name: "Test New repository client",
			args: args{
				config: &ClientConfig{
					BaseURL:    "http://artifactory.server.com",
					Username:   "admin",
					Password:   "password",
					Token:      "",
					AuthMethod: "",
					VerifySSL:  false,
					Client:     nil,
					Transport:  nil,
				}},
			want: Client{
				Client:    &http.Client{},
				Config:    &ClientConfig{},
				Transport: &http.Transport{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.args.config)
			if got.Client == nil {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRepositoryClient(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
		{
			name: "Test New Repository client without",
			want: &Client{
				Client:    &http.Client{},
				Config:    &ClientConfig{},
				Transport: &http.Transport{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("REPOSITORY_URL", "http://artifactory.server.com") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_USERNAME", "admin") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_PASSWORD", "password") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_TOKEN", "") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			got := NewRepositoryClient()
			if got.Client == nil {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
			if got.Config.BaseURL != "http://artifactory.server.com" {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
			if got.Config.Username != "admin" {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
			if got.Config.Password != "password" {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRepositoryClientFromTokem(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
		{
			name: "Test New Repository client without",
			want: &Client{
				Client:    &http.Client{},
				Config:    &ClientConfig{},
				Transport: &http.Transport{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("REPOSITORY_URL", "http://artifactory.server.com") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_TOKEN", "sometoken") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			got := NewRepositoryClient()
			if got.Client == nil {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
			if got.Config.BaseURL != "http://artifactory.server.com" {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
			if got.Config.Token != "sometoken" {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRepositoryClientErrorURLVariables(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
		{
			name: "Test New Repository client without",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("REPOSITORY_URL", "") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_TOKEN", "sometoken") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			got := NewRepositoryClient()
			if got != tt.want {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRepositoryClientErrorCredentialVariables(t *testing.T) {
	tests := []struct {
		name string
		want *Client
	}{
		{
			name: "Test New Repository client without",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("REPOSITORY_URL", "http://artifactory.server.com") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_USERNAME", "") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_PASSWORD", "") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			err = os.Setenv("REPOSITORY_TOKEN", "") //nolint
			if err != nil {
				t.Error("Error setting environment variable.")
			}
			got := NewRepositoryClient()
			if got != tt.want {
				t.Errorf("NewRepositoryClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
