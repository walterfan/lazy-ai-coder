package util

import (
	"crypto/tls"
	"net/http"
	"io"
)

func createHttpClient() (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: transport}, nil
}


func GetRemoteFileContent(url string) (string, error) {
	client, err := createHttpClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

}