package rss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (p *Parser) ParseURL(url string) (string, error) {
	body := map[string]string{
		"url": url,
	}
	byteBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	reader := bytes.NewReader(byteBody)
	response, err := http.Post(p.urlParser, "application/json", reader)
	if err != nil {
		return "", err
	}
	type Response struct {
		Text  string `json:"text"`
		Error string `json:"error"`
	}
	resp := Response{}
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return "", err
	} else if resp.Error != "" {
		return "", fmt.Errorf(resp.Error)
	} else {
		return resp.Text, nil
	}
}
