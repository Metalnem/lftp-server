package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

// Request represents single request for mirroring one FTP directory or a file.
type Request struct {
	Path     string `json:"path"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (request *Request) makeCmd() (*exec.Cmd, error) {
	if request.Path == "" {
		return nil, errors.New("No URL specified in a request")
	}

	url, err := url.Parse(request.Path)

	if err != nil {
		return nil, fmt.Errorf("Invalid URL: %s", request.Path)
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

func handler(w http.ResponseWriter, r *http.Request) {
	var request Request
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid request received")
		return
	}

	cmd, err := request.makeCmd()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	err = cmd.Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}

func main() {
	if _, err := exec.LookPath("lftp"); err != nil {
		log.Fatal("LFTP not found")
	}

	handler := http.HandlerFunc(handler)
	log.Fatal(http.ListenAndServe(":7800", handler))
}
