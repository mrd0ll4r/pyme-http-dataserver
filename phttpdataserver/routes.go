package phttpdataserver

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

func (s *httpDataServer) handlePut(w http.ResponseWriter, r *http.Request, p httprouter.Params) (int, error) {
	if s.testMode {
		io.Copy(ioutil.Discard, r.Body)
		return http.StatusOK, nil
	}

	path := p.ByName("path")
	if strings.Contains(path, "..") {
		return http.StatusForbidden, errors.New("disallowed path")
	}
	path = filepath.Join(s.wd, path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return http.StatusMethodNotAllowed, errors.New("file already exists")
	}

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return http.StatusForbidden, errors.Wrap(err, "unable to create directory")
	}

	f, err := os.Create(path)
	if err != nil {
		return http.StatusForbidden, errors.Wrap(err, "unable to create file")
	}
	defer f.Close()

	ms := r.URL.Query().Get("MirrorSource")
	if ms != "" {
		resp, err := http.Get(ms)
		if err != nil {
			return http.StatusNotFound, errors.Wrap(err, "unable to retrieve file")
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return http.StatusInternalServerError, errors.Wrap(err, "unable to write file")
		}
		return http.StatusOK, nil
	}

	_, err = io.Copy(f, r.Body)
	if err != nil {
		return http.StatusBadRequest, errors.Wrap(err, "unable to write file")
	}

	return http.StatusOK, nil
}

func (s *httpDataServer) handleGet(w http.ResponseWriter, r *http.Request, p httprouter.Params) (int, interface{}, error) {
	path := p.ByName("path")
	if strings.Contains(path, "..") {
		return http.StatusForbidden, nil, errors.New("disallowed path")
	}
	path = filepath.Join(s.wd, path)

	fs, err := os.Stat(path)
	if err != nil {
		return http.StatusForbidden, nil, errors.New("unable to access path")
	}

	if fs.IsDir() {
		return s.handleGetDirectory(w, r, p)
	}
	return s.handleGetFile(w, r, p)
}

func (s *httpDataServer) handleGetFile(w http.ResponseWriter, r *http.Request, p httprouter.Params) (int, interface{}, error) {
	path := p.ByName("path")
	if strings.Contains(path, "..") {
		return http.StatusForbidden, nil, errors.New("disallowed path")
	}
	path = filepath.Join(s.wd, path)

	f, err := os.Open(path)
	if err != nil {
		return http.StatusForbidden, nil, errors.Wrap(err, "unable to open file")
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream; charset=utf-8")
	w.WriteHeader(200)

	io.Copy(w, f)
	return 200, f, nil
}

func (s *httpDataServer) handleGetDirectory(w http.ResponseWriter, r *http.Request, p httprouter.Params) (int, interface{}, error) {
	path := p.ByName("path")
	if strings.Contains(path, "..") {
		return http.StatusForbidden, nil, errors.New("disallowed path")
	}
	path = filepath.Join(s.wd, path)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return http.StatusForbidden, nil, errors.Wrap(err, "unable to list directory")
	}
	names := make([]string, len(files))
	for i, f := range files {
		if f.IsDir() {
			names[i] = f.Name() + "/"
			continue
		}
		names[i] = f.Name()
	}

	return http.StatusOK, names, nil
}
