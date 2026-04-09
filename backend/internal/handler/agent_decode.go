package handler

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func decodeAgentJSON(_ http.ResponseWriter, r *http.Request, dst any) error {
	body, err := openPossiblyGzippedBody(r)
	if err != nil {
		return err
	}
	defer body.Close()

	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON object")
		}
		return err
	}

	return nil
}

func openPossiblyGzippedBody(r *http.Request) (io.ReadCloser, error) {
	if r.Body == nil {
		return io.NopCloser(strings.NewReader("")), nil
	}

	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("Content-Encoding")), "gzip") {
		return r.Body, nil
	}

	gr, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}

	return &multiCloser{
		Reader: gr,
		closers: []io.Closer{
			gr,
			r.Body,
		},
	}, nil
}

type multiCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiCloser) Close() error {
	var firstErr error
	for _, c := range m.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
