package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

// Request represents single request for mirroring one FTP directory or a file.
type Request struct {
	Path     string `json:"path"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Handler implements http.Handler interface and logs errors to custom log.Logger.
type Handler struct {
	Logger *log.Logger
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
		"-e", fmt.Sprintf("mirror '%s' && exit", url.Path),
		fmt.Sprintf("%s://%s", url.Scheme, url.Host),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, nil
}

func (handler *Handler) handle(w http.ResponseWriter, r *http.Request) error {
	var request Request
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		return fmt.Errorf("Invalid request received: %v", err)
	}

	cmd, err := request.makeCmd()

	if err != nil {
		return err
	}

	return cmd.Run()
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := handler.handle(w, r); err != nil {
		handler.Logger.Println(err)
	}
}

func main() {
	if _, err := exec.LookPath("lftp"); err != nil {
		log.Fatal("LFTP not found")
	}

	request := Request{
		Path:     "ftp://example.org/path",
		Username: "user",
		Password: "pass",
	}

	logger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)

	if err := encoder.Encode(request); err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second)
		resp, err := http.Post("http://localhost:7800/jsonrpc", "application/json", buffer)

		if err != nil {
			logger.Println(err)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Println(resp.Status)
		}
	}()

	http.Handle("/jsonrpc", &Handler{Logger: logger})
	log.Fatal(http.ListenAndServe(":7800", nil))
}
