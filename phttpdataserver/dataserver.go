package phttpdataserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

// HTTPDataServer is a data server for PYME that uses HTTP.
type HTTPDataServer interface {
	// Run runs the server.
	// This blocks.
	Run() error
}

type httpDataServer struct {
	r    *httprouter.Router
	wd   string
	port int
}

// New creates a new HTTPDataServer using the port and working directory
// specified.
func New(port int, workingDirectory string) HTTPDataServer {
	toReturn := &httpDataServer{
		r:    httprouter.New(),
		wd:   workingDirectory,
		port: port,
	}

	toReturn.r.PUT("/*path", toReturn.makeHandler(noResultHandler(toReturn.handlePut)))
	toReturn.r.GET("/*path", toReturn.makeHandler(toReturn.handleGet))

	return toReturn
}

func (s *httpDataServer) Run() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	log.Printf("Listening on %s...\n", addr)
	log.Printf("Serving %s...\n", s.wd)
	return http.ListenAndServe(addr, s.r) // this blocks
}

// ResponseFunc is the type of function that handles an API request and returns
// an HTTP status code, an optional response to be embedded and an optional error.
type ResponseFunc func(http.ResponseWriter, *http.Request, httprouter.Params) (status int, result interface{}, err error)

// NoResultResponseFunc is the type of function that handles an API request and
// returns an HTTP status code and an optional error.
type NoResultResponseFunc func(http.ResponseWriter, *http.Request, httprouter.Params) (status int, err error)

// ErrInternalServerError is the error used for recovered API calls.
var ErrInternalServerError = errors.New("internal server error")

type response struct {
	Ok     bool        `json:"ok"`
	Error  string      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func (s *httpDataServer) makeHandler(inner ResponseFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		//resp := response{}
		handler := logHandler(recoverHandler(inner))

		status, result, err := handler(w, r, p)
		/*if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Ok = true
		}
		if result != nil {
			resp.Result = result
		}*/

		// This whole bit is a "type switch" and not nice. I would solve
		// it differently if we used the standardized response.

		if result != nil {
			switch result.(type) {
			case *os.File:
				// handled by handleGetFile
				// This is pretty dirty.
				break
			default:
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(status)
				err = json.NewEncoder(w).Encode(result)
				if err != nil {
					log.Printf("phttpdataserver: unable to encode JSON: %s\n", err)
				}
			}
		}

		if result != nil {
			switch result.(type) {
			case *os.File:
				w.Header().Set("Content-Type", "application/octet-stream; charset=utf-8")
			default:
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
			}
		}
	}
}

func logHandler(inner ResponseFunc) ResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) (int, interface{}, error) {
		before := time.Now()

		status, result, err := inner(w, r, p)
		delta := time.Since(before)

		if err != nil {
			log.Printf("%d(%s) %s %s %s %s", status, err.Error(), delta.String(), r.RemoteAddr, r.Method, r.URL.EscapedPath())
		} else {
			log.Printf("%d %s %s %s %s", status, delta.String(), r.RemoteAddr, r.Method, r.URL.EscapedPath())
		}
		return status, result, err
	}
}

func recoverHandler(inner ResponseFunc) ResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) (status int, result interface{}, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Println("API: recovered:", rec)
				status = http.StatusInternalServerError
				result = nil
				err = ErrInternalServerError
			}
		}()

		status, result, err = inner(w, r, p)
		return
	}
}

func noResultHandler(inner NoResultResponseFunc) ResponseFunc {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) (int, interface{}, error) {
		status, err := inner(w, r, p)

		return status, nil, err
	}
}
