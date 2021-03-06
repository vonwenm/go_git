// The base functions for making requests and returning the JSON from them
package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// New type to make the auth token more explicit
type OAuthToken string

// Request struct is used to construct the query URIs
type Request struct {
	BaseURI string
	Query   string
	Page    int
	PerPage int
}

// query param should have no leading or trailing slashes
func NewRequest(query string) *Request {
	return &Request{
		BaseURI: "https://api.github.com",
		Query:   query,
		Page:    1,
		PerPage: 100,
	}
}

// Constructs an API request with page and per_page options
// note that events endpoints currently only allow per_page of 30
// all other requests can have a per_page of up to 100
func (r *Request) String() string {
	return fmt.Sprintf("%s/%s?page=%v&per_page=%v",
		r.BaseURI, r.Query, r.Page, r.PerPage)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Generalized API request function that iterates through subsequent pages
// token:				A string containing a Github OAuth token
// baseRequest:	The API args only, with no leading or trailing slashes
//							eg: "users/octocat/repos"
func APIRequest(request *Request, token *OAuthToken) []map[string]interface{} {
	done := false
	fin := make([]map[string]interface{}, 0)
	client := &http.Client{Timeout: 5 * time.Second}

	for !done {
		// Create a request with the OAuth token in the header
		req, err := http.NewRequest("GET", request.String(), nil)
		check(err)

		if *token != "" {
			req.Header.Add("Authorization", fmt.Sprintf("token %s", *token))
		}

		// Send the request and read the response
		resp, err := client.Do(req)
		check(err)
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		check(err)

		// Parse the response to JSON
		// The API will return either []map[string]interface{} or a single
		// map[string]interface{}, if we get a single item we wrap it in a slice
		// to make the return values consistent across the board
		var js []map[string]interface{}
		err = json.Unmarshal(body, &js)
		if err != nil {
			// If unmarhsaling failed the return value was a single JSON object
			obj := make(map[string]interface{})
			err = json.Unmarshal(body, &obj)
			js = append(js, obj)
		}

		// If that page was less than the page limit we are done
		if len(js) < request.PerPage {
			done = true
		}

		// Append the unmarshaled JSON from this page to the final array
		for _, item := range js {
			fin = append(fin, item)
		}

		request.Page++
	}

	return fin
}

// Read the rate limit, currently 5000 requests per hour when auth'd
// and 60 when not
func RateLimit(token *OAuthToken) int {
	req := NewRequest("rate_limit")
	js := APIRequest(req, token)
	rate := js[0]["rate"].(map[string]interface{})
	return int(rate["limit"].(float64))
}

// Reads the remaining rate limit for the token, with an empty string it
// returns the remaining unauth'd rate limit for the IP
func RateLimitRemaining(token *OAuthToken) int {
	req := NewRequest("rate_limit")
	js := APIRequest(req, token)
	rate := js[0]["rate"].(map[string]interface{})
	return int(rate["remaining"].(float64))
}
