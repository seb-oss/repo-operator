package repository

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
)

// ClientConfig is the configuration for an REPOSITORY Client
type ClientConfig struct {
	BaseURL    string
	Username   string
	Password   string
	Token      string
	AuthMethod string
	VerifySSL  bool
	Client     *http.Client
	Transport  *http.Transport
}

// Client is a client for interacting with REPOSITORY
type Client struct {
	Client    *http.Client
	Config    *ClientConfig
	Transport *http.Transport
	rt        rtInterface
}

// NewRepositoryClient : Create new repository client from environment variable
func NewRepositoryClient() *Client {
	config, err := clientConfigFrom("environment")
	if err != nil {
		return nil
	}

	client := NewClient(config)

	return &client
}

// NewClient returns a new REPOSITORY Client with the provided ClientConfig
func NewClient(config *ClientConfig) Client {
	verifySSL := func() bool {
		return !config.VerifySSL
	}
	if config.Transport == nil {
		config.Transport = &http.Transport{}
	}
	config.Transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: verifySSL()}
	if config.Client == nil {
		config.Client = &http.Client{}
	}
	config.Client.Transport = config.Transport
	return Client{Client: config.Client, Config: config, Transport: config.Transport, rt: RTFactory{}}
}

func clientConfigFrom(from string) (*ClientConfig, error) {
	conf := ClientConfig{}
	switch from {
	case "environment":
		if os.Getenv("REPOSITORY_URL") == "" {
			return nil, errors.New("You must set the environment variable REPOSITORY_URL")
		}

		conf.BaseURL = os.Getenv("REPOSITORY_URL")
		if os.Getenv("REPOSITORY_TOKEN") == "" {
			if os.Getenv("REPOSITORY_USERNAME") == "" || os.Getenv("REPOSITORY_PASSWORD") == "" {
				return nil, errors.New("You must set the environment variables REPOSITORY_USERNAME & REPOSITORY_PASSWORD")
			}

			conf.AuthMethod = "basic"
		} else {
			conf.AuthMethod = "token"
		}
	}
	if conf.AuthMethod == "token" {
		conf.Token = os.Getenv("REPOSITORY_TOKEN")
	} else {
		conf.Username = os.Getenv("REPOSITORY_USERNAME")
		conf.Password = os.Getenv("REPOSITORY_PASSWORD")
	}
	return &conf, nil
}
