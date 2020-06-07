# mock-http

Mock http server response if matching with mock definitions.
Reverse proxy to original location if no matching with mock definitions.

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
➜  mock-http git:(master) cat ~/.mock-http/definitions/fetch_kyc.mock
{
    "url": "http://example.com:5000/api/test",
    "method": "POST",
    "response": {
        "default": {
            "ok": "ok"
        },
        "{{ and ( eq .Body.test \"test1\" )  (true ) }}": {
            "ok": "test1"
        },
        "{{ and ( eq .Body.test \"test2\" )  (true ) }}": {
            "ok": "test2"
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

	req, err := http.NewRequest("GET", "http://example.com:5000/api/test", bytes.NewReader([]byte{}))
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
