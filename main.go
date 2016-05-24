package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"golang.org/x/crypto/bcrypt"
)

var (
	rpcListenPort = flag.Int("rpc-listen-port", 7800, "Specify a port number for JSON-RPC server to listen to. Possible values: 1024-65535")
	rpcSecret     = flag.String("rpc-secret", "", "Set RPC secret authorization token (required)")

	// Info is used for logging information.
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Error is used for logging errors.
	Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	errMissingURL           = errors.New("No URL specified in a request")
	errProtocolMismatch     = errors.New("Only FTP downloads are supported")
	errInvalidRequestFormat = errors.New("Invalid request format")
	errTokenMismatch        = errors.New("Secret token does not match")
	errUnauthorized         = errors.New("Missing or invalid credentials")
)

// Request represents single request for mirroring one FTP directory or a file.
type Request struct {
	Path     string `json:"path"`
	Username string `json:"username"`
	Password string `json:"password"`
	Secret   string `json:"secret"`
}

// Response represents response to a client with ID for a created job or error message in case of error.
type Response struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Handler implements http.Handler interface and processes download requests sequentially.
type Handler struct {
	Jobs        chan *Job
	HashedToken []byte
}

// Job is single download request with associated ID and LFTP command.
type Job struct {
	ID      string
	Command *exec.Cmd
}

func (request *Request) extractURL() (*url.URL, error) {
	if request.Path == "" {
		return nil, errMissingURL
	}

	url, err := url.Parse(request.Path)

	if err != nil {
		return nil, fmt.Errorf("Invalid URL: %s", request.Path)
	}

	if url.Scheme != "ftp" {
		return nil, errProtocolMismatch
	}

	return url, nil
}

func makeLftpCmd(path string) string {
	if path == "" {
		return "mirror && exit"
	}

	lftpCmd := "pget"

	if strings.HasSuffix(path, "/") {
		lftpCmd = "mirror"
	}

	escaped := strings.Replace(path, "\"", "\\\"", -1)
	return fmt.Sprintf("%s \"%s\" && exit", lftpCmd, escaped)
}

func makeCmd(url *url.URL, username, password string) *exec.Cmd {
	lftpCmd := makeLftpCmd(url.Path)
	var args []string

	if username != "" && password != "" {
		args = []string{"--user", username, "--password", password, "-e", lftpCmd, url.Host}
	} else {
		args = []string{"-e", lftpCmd, url.Host}
	}

	cmd := exec.Command("lftp", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func generateID() string {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		panic("Random number generator failed")
	}

	return base64.StdEncoding.EncodeToString(b)
}

func (handler *Handler) serveHTTP(w http.ResponseWriter, r *http.Request, id string) error {
	var request Request
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		return errInvalidRequestFormat
	}

	if err := bcrypt.CompareHashAndPassword(handler.HashedToken, []byte(request.Secret)); err != nil {
		return errTokenMismatch
	}

	Info.Printf("Download request %s has URL %s\n", id, request.Path)
	url, err := request.extractURL()

	if err != nil {
		return err
	}

	conn, err := ftp.DialTimeout(fmt.Sprintf("%s:21", url.Host), 5*time.Second)

	if err != nil {
		return fmt.Errorf("Unable to connect to FTP server at %s", url.Host)
	}

	if request.Username != "" && request.Password != "" {
		err = conn.Login(request.Username, request.Password)
	} else {
		err = conn.Login("anonymous", "anonymous")
	}

	if err != nil {
		return errUnauthorized
	}

	conn.Logout()
	json.NewEncoder(w).Encode(Response{ID: id})

	cmd := makeCmd(url, request.Username, request.Password)
	job := Job{ID: id, Command: cmd}

	go func() {
		handler.Jobs <- &job
	}()

	return nil
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := generateID()
	Info.Printf("Received download request %s from %s\n", id, r.RemoteAddr)

	if err := handler.serveHTTP(w, r, id); err != nil {
		Error.Printf("Invalid request %s: %s\n", id, err)
		status := http.StatusBadRequest

		if err == errUnauthorized {
			status = http.StatusUnauthorized
		}

		w.WriteHeader(status)
		json.NewEncoder(w).Encode(Response{Message: err.Error()})
	}
}

func (handler *Handler) worker() {
	for job := range handler.Jobs {
		Info.Printf("Begin LFTP output for request %s", job.ID)
		err := job.Command.Run()
		Info.Printf("End LFTP output for request %s", job.ID)

		if err != nil {
			Error.Printf("Failed to execute request %s with error: %v\n", job.ID, err)
		} else {
			Info.Printf("Request %s completed", job.ID)
		}
	}
}

func main() {
	flag.Parse()

	if (*rpcListenPort < 1024 || *rpcListenPort > 65535) || *rpcSecret == "" {
		flag.Usage()
		os.Exit(1)
	}

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(*rpcSecret), bcrypt.DefaultCost)

	if err != nil {
		log.Fatal("bcrypt failed to generate hashed token")
	}

	if _, err := exec.LookPath("lftp"); err != nil {
		log.Fatal("LFTP not found")
	}

	handler := &Handler{
		Jobs:        make(chan *Job, 10),
		HashedToken: hashedToken,
	}

	http.Handle("/jsonrpc", handler)
	go handler.worker()

	Info.Printf("Starting LFTP server on port %d\n", *rpcListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *rpcListenPort), nil))
}
