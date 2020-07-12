# mock-http

Mock http server response if matching with mock definitions.
Reverse proxy to original location if no matching with mock definitions.

## How to install

```sh
// install or upgrade
go get -u github.com/aaabhilash97/mock-http
```

OR
Download from https://github.com/aaabhilash97/mock-http/releases

## Usage

```bash
Usage of ./mock-http:
  -address string
    	Address  Ex: 3000, 0.0.0.0:3000 (default "127.0.0.1:3000")
  -definitions string
    	Mock definitions location (default "~/.mock-http/definitions")
```

## Mock definition example

```bash
➜  mock-http git:(master) cat ~/.mock-http/definitions/example1.mock
{
    "url": "http://example.com:5000/api/test",
    "method": "POST",
    "response": {
        "default": {
            "data": {
                "body": {
                    "status-code": "102",
                    "result": {
                        "name": "ABHILASH KM"
                    }
                }
            },
            "header": {
                "Code": "0000"
            }
        },
        "{{ and (eq .RawBody \"qqqq\") (true) }}": {
            "data": {
                "body": {
                    "status-code": "100",
                    "result": {
                        "name": "ABHILASH KM"
                    }
                }
            },
            "header": {
                "Code": "0000"
            }
        },
        "{{if .Body.param}}{{ and (eq .Body.param \"qqqq\") (true) }}{{end}}": {
            "data": {
                "body": {
                    "status-code": "102",
                    "result": {
                        "name": "ABHILASH KM"
                    }
                }
            },
            "header": {
                "Code": "0000"
            }
        },
        "{{if .Query.param}}{{ and (eq .Query.param \"csrpm7372k\") (true) }}{{end}}": {
            "data": {
                "body": {
                    "status-code": "102",
                    "result": {
                        "name": "ABHILASH KM"
                    }
                }
            }
        }
    },
    "content_type": "application/json"
}
```

## Example Go client using mock-http as proxy

```go
package main
import (
    "bytes"
	"io/ioutil"
	"net/http"
	"log"
)
func main() {
	var netTransport = &http.Transport{
		Proxy:               "http://127.0.0.1:3000",
	}

	var netClient = &http.Client{
		Transport: netTransport,
	}

	req, err := http.NewRequest("POST", "http://example.com:5000/api/test", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ioutil.ReadAll(resp.Body))
}

```
