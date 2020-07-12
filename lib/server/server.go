package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	applicationJson string = "application/json"
)

type Options struct {
	Address             string
	DefinitionsLocation string
	Debug               bool
}

type MockDefinition struct {
	Url         string                 `json:"url"`
	Response    map[string]interface{} `json:"response"`
	ContentType string                 `json:"content_type"`
	Method      string                 `json:"method"`
}

type reqValuesModel struct {
	Body   map[string]interface{}
	Query  map[string]string
	Header map[string]string
}

func StartServer(opt Options) error {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defFiles, err := ioutil.ReadDir(opt.DefinitionsLocation)
		if err != nil {
			if opt.Debug {
				log.Print(err)
			}
		}

		query := make(map[string]string)
		for key, value := range req.URL.Query() {
			if len(value) > 0 {
				query[key] = value[0]
			}
		}
		header := make(map[string]string)
		for key, value := range req.Header {
			if len(value) > 0 {
				header[key] = value[0]
			}
		}

		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err)
			return
		}
		body := make(map[string]interface{})
		if len(bodyBytes) > 0 {
			err = json.Unmarshal(bodyBytes, &body)
			if err != nil {
				log.Println(err)
			}
		}

		reqValue := reqValuesModel{
			Body:   body,
			Query:  query,
			Header: header,
		}

		// Check for mockable
		for _, defFile := range defFiles {
			if !defFile.IsDir() {
				file, err := ioutil.ReadFile(path.Join(opt.DefinitionsLocation, defFile.Name()))
				if err != nil {
					log.Print(err)
					continue
				}
				mock := MockDefinition{}
				err = json.Unmarshal(file, &mock)
				if err != nil {
					log.Print("JSON:", err)
					continue
				}
				if matchRequestWithMock(mock, req) {
					if err := cc(mock, reqValue, w); err != nil {
						continue
					} else {
						return
					}
				}
			}
		}
		if req.URL.Host == "" {
			fmt.Fprintf(w, "No mock definitions matching")
			return
		}

		if opt.Debug {
			log.Println(fmt.Sprintf("Calling %s://%s", req.URL.Scheme, req.URL.Host))
		}
		// Reproxying request to original host
		remote, err := url.Parse(fmt.Sprintf("%s://%s", req.URL.Scheme, req.URL.Host))
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Failed to proxy request")
			return
		}
		dialer := &net.Dialer{
			DualStack: true,
		}
		http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		req.Host = remote.Host
		{

			var proxyReqBody io.Reader = bytes.NewReader(bodyBytes)
			var ok bool
			req.Body, ok = proxyReqBody.(io.ReadCloser)
			if !ok && body != nil {
				req.Body = ioutil.NopCloser(proxyReqBody)
			}
		}

		proxy.ServeHTTP(w, req)
	})

	log.Println(fmt.Sprintf("Running in %s", opt.Address))
	err := http.ListenAndServe(opt.Address, nil)
	// err := http.ListenAndServeTLS(opt.Address, "./server.crt", "./server.key", nil)
	if err != nil {
		if opt.Debug {
			log.Print(err)
		}
	}
	return err
}

func matchRequestWithMock(mock MockDefinition, req *http.Request) bool {
	if mock.Url == fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path) && req.Method == mock.Method {
		return true
	}
	return false
}

func cc(mock MockDefinition, reqValues reqValuesModel, w http.ResponseWriter) error {
	if len(mock.ContentType) > 0 {
		w.Header().Set("Content-Type", mock.ContentType)
	}

	for key, value := range mock.Response {
		if key == "default" {
			continue
		}
		t := template.Must(template.New("letter").Parse(key))
		var keyVal bytes.Buffer
		err := t.Execute(&keyVal, reqValues)
		if err != nil {
			log.Println(err)
			return err
		}
		if b, _ := strconv.ParseBool(strings.TrimSpace(keyVal.String())); b {
			if err := sendResponse(value, mock, w); err != nil {
				continue
			} else {
				return nil
			}
		}
	}
	if value, ok := mock.Response["default"]; ok {
		err := sendResponse(value, mock, w)
		return err
	}
	return errors.New("No mock matching")
}

func sendResponse(value interface{}, mock MockDefinition, w http.ResponseWriter) error {
	if _, ok := value.(map[string]interface{}); ok {
		if len(mock.ContentType) == 0 {
			w.Header().Set("Content-Type", applicationJson)
		}
		responseBytes, err := json.Marshal(value)
		if err != nil {
			log.Println(err)
			return err
		}

		w.Write(responseBytes)
		return nil
	}
	fmt.Fprintf(w, "%s", mock.Response)
	return nil
}
