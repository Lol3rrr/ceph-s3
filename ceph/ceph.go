package ceph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Ceph struct {
	baseurl string
	token   string
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func CephAuth(baseurl string, username string, password string) (Ceph, error) {
	jsonBody, err := json.Marshal(AuthRequest{
		Username: username,
		Password: password,
	})
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/auth", baseurl), bodyReader)
	if err != nil {
		return Ceph{}, nil
	}

	req.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return Ceph{}, err
	}

	defer res.Body.Close()
	var response AuthResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return Ceph{}, err
	}

	token := response.Token
	if len(token) == 0 {
		return Ceph{}, errors.New("Got empty Token Response")
	}

	return Ceph{token: token, baseurl: baseurl}, nil
}

func (ceph *Ceph) RgwUsers() ([]string, error) {
	jsonBody := []byte{}
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/rgw/user", ceph.baseurl), bodyReader)
	if err != nil {
		return []string{}, nil
	}

	req.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ceph.token))

	fmt.Printf("Token: '%s'", ceph.token)

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}

	defer res.Body.Close()
	var response []string
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return []string{}, err
	}

	return response, nil
}

type SetKeyRequest struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	KeyType   string `json:"key_type"`
}

type SetKeyResponse struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

func (ceph *Ceph) AddKey(uid string, access_key string, secret_key string) (SetKeyResponse, error) {
	jsonBody, err := json.Marshal(SetKeyRequest{
		AccessKey: access_key,
		SecretKey: secret_key,
		KeyType:   "s3",
	})
	bodyReader := bytes.NewReader(jsonBody)
	url := fmt.Sprintf("%s/api/rgw/user/%s/key", ceph.baseurl, uid)
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return SetKeyResponse{}, nil
	}

	req.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ceph.token))

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return SetKeyResponse{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return SetKeyResponse{}, err
		}

		return SetKeyResponse{}, errors.New(fmt.Sprintf("Non 200 Status Code: %d\nURL: %s\n%s", res.StatusCode, url, string(b)))
	}

	var response []SetKeyResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return SetKeyResponse{}, err
	}

	for _, key := range response {
		if key.SecretKey == secret_key {
			return key, nil
		}
	}

	return SetKeyResponse{}, errors.New("Could not find new key in Key Response")
}

func (ceph *Ceph) RemoveKey(uid string, access_key string) error {
	jsonBody := []byte{}
	bodyReader := bytes.NewReader(jsonBody)
	url := fmt.Sprintf("%s/api/rgw/user/%s/key?access_key=%s", ceph.baseurl, uid, access_key)
	req, err := http.NewRequest(http.MethodDelete, url, bodyReader)
	if err != nil {
		return nil
	}

	req.Header.Set("Accept", "application/vnd.ceph.api.v1.0+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ceph.token))

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 202 && res.StatusCode != 204 {
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return errors.New(fmt.Sprintf("Non 200 Status Code: %d\nURL: %s\n%s", res.StatusCode, url, string(b)))
	}

	return nil
}
