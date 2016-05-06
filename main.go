package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
)

// Request represents single request for mirroring one FTP directory or a file.
type Request struct {
	URL      string
	Username string
	Password string
}

func (request *Request) makeCmd() (*exec.Cmd, error) {
	if request.URL == "" {
		return nil, errors.New("No URL specified in a request")
	}

	url, err := url.Parse(request.URL)

	if err != nil {
		return nil, fmt.Errorf("Invalid URL: %s", request.URL)
	}

	cmd := exec.Command(
		"lftp",
		"-u", fmt.Sprintf("%s,%s", request.Username, request.Password),
		"-e", fmt.Sprintf("mirror %s && exit", url.Path),
		fmt.Sprintf("%s://%s", url.Scheme, url.Host),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, nil
}

func main() {
	if _, err := exec.LookPath("lftp"); err != nil {
		log.Fatal("LFTP not found")
	}

	request := Request{
		URL:      "ftp://example.org/",
		Username: "username",
		Password: "password",
	}

	cmd, err := request.makeCmd()

	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}
