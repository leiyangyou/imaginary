package main

import (
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

const maxMemory int64 = 1024 * 1024 * 64

func readBody(r *http.Request) ([]byte, error) {
	var err error
	var buf []byte

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/") {
		err = r.ParseMultipartForm(maxMemory)
		if err != nil {
			return nil, err
		}

		buf, err = readFormPayload(r)
	} else {
		buf, err = ioutil.ReadAll(r.Body)
	}

	return buf, err
}

func readFormPayload(r *http.Request) ([]byte, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, NewError("Empty payload", BAD_REQUEST)
	}

	return buf, err
}

func readPayload(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Method != "POST" {
		return nil, ErrorReply(w, "Method not allowed for this endpoint", NOT_ALLOWED)
	}

	buf, err := readBody(r)
	if err != nil {
		return nil, ErrorReply(w, "Cannot read the payload: "+err.Error(), BAD_REQUEST)
	}

	return buf, nil
}

func readLocalImage(w http.ResponseWriter, r *http.Request, mountPath string) ([]byte, error) {
	file := r.URL.Query().Get("file")
	if file == "" {
		return nil, ErrorReply(w, "Missing required param: file", BAD_REQUEST)
	}

	file = path.Clean(path.Join(mountPath, file))
	if strings.HasPrefix(file, mountPath) == false {
		return nil, ErrorReply(w, "Invalid file path", BAD_REQUEST)
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, ErrorReply(w, "Invalid file path", BAD_REQUEST)
	}

	return buf, nil
}
