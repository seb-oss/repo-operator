package repository

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var statusCodeStateList = make(map[int]string)

// Request represents an artifactory http request
type Request struct {
	Verb        string
	Path        string
	ContentType string
	Accept      string
	QueryParams map[string]string
	Body        io.Reader
}

// Get performs an http GET to artifactory
func Get(c *Client, path string, options map[string]string) ([]byte, int, string, error) {
	r, err := makeRequest(c, "GET", path, options, nil)
	if err != nil {
		var data bytes.Buffer
		return data.Bytes(), 500, statusInternalServerErrorState, err
	}

	return parseResponse(r)
}

// Put performs an http PUT to artifactory
func Put(c *Client, path string, data []byte, options map[string]string) ([]byte, int, string, error) {
	body := bytes.NewReader(data)
	r, err := makeRequest(c, "PUT", path, options, body)
	if err != nil {
		var data bytes.Buffer
		return data.Bytes(), 500, statusInternalServerErrorState, err
	}

	return parseResponse(r)
}

// Delete performs an http DELETE to artifactory
func Delete(c *Client, path string) (int, string, error) {
	var code = 200
	r, err := makeRequest(c, "DELETE", path, make(map[string]string), nil)
	if err != nil {
		code = 500
		return code, statusInternalServerErrorState, err
	}
	_, _, _, err = parseResponse(r)

	return code, statusOKState, err
}

func makeRequest(c *Client, method string, path string, options map[string]string, body io.Reader) (*http.Response, error) {
	qs := url.Values{}
	var contentType string
	for q, p := range options {
		if q == "content-type" {
			contentType = p
			delete(options, q)
		} else {
			qs.Add(q, p)
		}
	}

	baseReqPath := strings.TrimSuffix(c.Config.BaseURL, "/") + path
	u, err := url.Parse(baseReqPath)
	if err != nil {
		return nil, err
	}
	if len(options) != 0 {
		u.RawQuery = qs.Encode()
	}
	buf := new(bytes.Buffer)
	if body != nil {
		_, _ = buf.ReadFrom(body)
	}
	req, _ := http.NewRequest(method, u.String(), bytes.NewReader(buf.Bytes()))
	if body != nil {
		h := sha1.New()
		_, _ = h.Write(buf.Bytes())
		chkSum := h.Sum(nil)
		req.Header.Add("X-Checksum-Sha1", fmt.Sprintf("%x", chkSum))
	}
	req.Header.Add("user-agent", "artifactory-go."+VERSION)
	req.Header.Add("X-Result-Detail", "info, properties")
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	if c.Config.AuthMethod == "basic" {
		req.SetBasicAuth(c.Config.Username, c.Config.Password)
	} else {
		req.Header.Add("X-JFrog-Art-Api", c.Config.Token)
	}
	if os.Getenv("REPOSITORY_DEBUG") != "" {
		log.Info("Headers: %#v", req.Header)
		if len(buf.Bytes()) > 0 {
			log.Info("Body: %#v", buf.String())
		}
	}

	r, err := c.Client.Do(req)

	return r, err
}

func parseResponse(r *http.Response) ([]byte, int, string, error) {
	defer func() { _ = r.Body.Close() }()
	data, err := ioutil.ReadAll(r.Body)
	if (r.StatusCode < 200 || r.StatusCode > 299) && r.StatusCode != 400 {
		var ej ErrorsJSON
		uerr := json.Unmarshal(data, &ej)
		if uerr != nil {
			emsg := fmt.Sprintf("Unable to parse error json. Non-2xx code returned: %d. Message follows:\n%s", r.StatusCode, string(data))
			return data, r.StatusCode, statusCodeStateList[r.StatusCode], errors.New(emsg)
		}
		// here we catch the {"error":"foo"} oddity in things like security/apiKey
		if ej.Error != "" {
			return data, r.StatusCode, statusCodeStateList[r.StatusCode], errors.New(ej.Error)
		}
		var emsgs []string
		for _, i := range ej.Errors {
			emsgs = append(emsgs, i.Message)
		}
		return data, r.StatusCode, statusCodeStateList[r.StatusCode], errors.New(strings.Join(emsgs, "\n"))
	}
	return data, r.StatusCode, statusCodeStateList[r.StatusCode], err
}
