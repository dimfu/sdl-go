package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func MakeGetRequest(
	url string,
	queryBuilder func(q url.Values),
	result interface{},
) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	queryBuilder(q)

	req.URL.RawQuery = q.Encode()
	fmt.Println(req.URL.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
