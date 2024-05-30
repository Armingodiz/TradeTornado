package lib

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
)

type HttpEncoding string

const (
	Json HttpEncoding = "json"
	XML  HttpEncoding = "xml"
)

type HttpBody struct {
	Encoding HttpEncoding
	Body     interface{}
}

func (h *HttpBody) GetBody() (*bytes.Buffer, error) {
	switch h.Encoding {
	case Json:
		{
			data, err := json.Marshal(h.Body)
			if err != nil {
				logrus.Println("Error marshaling requset body:", err)
				return nil, err
			}
			return bytes.NewBuffer(data), nil
		}
	case XML:
		{
			data, err := xml.Marshal(h.Body)
			if err != nil {
				logrus.Println("Error marshaling requset body:", err)
				return nil, err
			}
			return bytes.NewBuffer(data), nil
		}
	default:
		return nil, errors.New("invalid encoding")
	}
}

func SendHttpRequest(url, method string, body *HttpBody, headers map[string]string) (*http.Response, error) {
	var req *http.Request
	var err error
	if body != nil {
		reqBody, err := body.GetBody()
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, reqBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	return resp, nil
}
