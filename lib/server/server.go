package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"time"
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

func StartServer(opt Options) error {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defFiles, err := ioutil.ReadDir(opt.DefinitionsLocation)
		if err != nil {
			if opt.Debug {
				log.Print(err)
			}
		}
		// Check for mockable
		for _, defFile := range defFiles {
			if !defFile.IsDir() {
				file, err := ioutil.ReadFile(path.Join(opt.DefinitionsLocation, defFile.Name()))
				if err != nil {
					if opt.Debug {
						log.Print(err)
					}
					continue
				}
				mock := MockDefinition{}
				err = json.Unmarshal(file, &mock)
				if err != nil {
					if opt.Debug {
						log.Print(err)
					}
					continue
				}
				if matchRequestWithMock(mock, req) {
					if err := cc(mock, req, w); err != nil {
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

		// Reproxying request to original host
		remote, err := url.Parse(fmt.Sprintf("%s://%s", req.URL.Scheme, req.URL.Host))
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, "Failed to proxy request")
			return
		}
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}
		http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		req.Host = remote.Host
		proxy.ServeHTTP(w, req)
	})

	err := http.ListenAndServe(opt.Address, nil)
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

func cc(mock MockDefinition, req *http.Request, w http.ResponseWriter) error {
	body := make(map[string]interface{})
	// query := req.URL.Query()
	// headers := req.Header
	type reqValuesModel struct {
		Body map[string]interface{}
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		return err
	}
	reqValues := reqValuesModel{
		Body: body,
	}

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
			return err
		}
		if b, _ := strconv.ParseBool(keyVal.String()); b {
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
			return err
		}
		w.Write(responseBytes)
		return nil
	}
	fmt.Fprintf(w, "%s", mock.Response)
	return nil
}
