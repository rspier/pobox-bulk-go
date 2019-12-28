// Package whitelabel is a client for the Pobox Whitelabel Bulk Route Management API.
package whitelabel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Client implements the API.
type Client struct {
	User, Pass string // GUID, API Key
}

// Route represents a single route.
type Route struct {
	// Fwd is the target forwarding address.
	Fwd string `json:"fwd"`
}

// Routes maps aliases to their respective Route.
type Routes map[string]*Route

// Counts returns a mapping of domain to number of aliases.
// TODO: ask why the API call returns a string instead of a int.
type Counts map[string]string

const base = "https://api.pobox.com/v2/"

// String() implements the Stringer interface for *Route so to simplfiy displaying them, especially when they're nested in Routes.
func (r *Route) String() string {
	if r == nil {
		return "nil"
	}
	return fmt.Sprintf("%+v", *r)
}

// newRequestWithContext creates a new request and sets the authentication information on it.
func (c *Client) newRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, base+url, body)
	if err != nil {
		return nil, err
	}
	r.SetBasicAuth(c.User, c.Pass)
	return r, nil
}

// do is a helper to simplify API requests.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	r, err := c.newRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	// read resp.Body here so we don't have to worry about conditionally closing
	// it in the caller if it's not nil.
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http response %v: %s", resp.StatusCode, string(b))
	}
	return b, nil
}

// GetRoute gets the route for a single alias.
func (c *Client) GetRoute(ctx context.Context, alias string) (*Route, error) {
	var route Route
	body, err := c.do(ctx, "GET", "route/"+alias, nil)
	err = json.Unmarshal(body, &route)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling route %s: %w", alias, err)
	}
	return &route, nil
}

// SetRoute creates or updates an alias.
func (c *Client) SetRoute(ctx context.Context, alias string, fwd string) error {
	var b []byte
	var err error
	if fwd != "" {
		b, err = json.Marshal(Route{Fwd: fwd})
		if err != nil {
			return fmt.Errorf("marhsalling route %q JSON: %w", fwd, err)
		}
	}
	_, err = c.do(ctx, "PUT", "route/"+alias, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("PUTting alias %s: %w", alias, err)

	}
	return nil
}

// DeleteRoute deletes an alias.
func (c *Client) DeleteRoute(ctx context.Context, alias string) error {
	return c.SetRoute(ctx, alias, "")
}

// GetRoutes retrieves all authorized routes.
func (c *Client) GetRoutes(ctx context.Context) (Routes, error) {
	var routes Routes
	body, err := c.do(ctx, "GET", "routes", nil)
	err = json.Unmarshal(body, &routes)
	if err != nil {
		return nil, err
	}
	return routes, nil
}

// SetRoutes sets routes for multiple aliases at once.  To delete the routing for an alias specify nil as the Route.
func (c *Client) SetRoutes(ctx context.Context, rs Routes) (Routes, error) {
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	err := e.Encode(rs)
	if err != nil {
		return nil, fmt.Errorf("marshalling routes to JSON: %w", err)
	}
	body, err := c.do(ctx, "POST", "routes", &buf)
	if err != nil {
		return nil, fmt.Errorf("POSTting routes: %w", err)
	}

	var routes Routes
	err = json.Unmarshal(body, &routes)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling routes from JSON: %w", err)
	}
	return routes, nil
}

// CountRoutes returns the count of routes for the authorized domains.
func (c *Client) CountRoutes(ctx context.Context) (*Counts, error) {
	body, err := c.do(ctx, "GET", "routes/count", nil)
	if err != nil {
		return nil, fmt.Errorf("GETting routes: %w", err)
	}

	var counts Counts
	err = json.Unmarshal(body, &counts)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling counts from JSON: %w", err)
	}
	return &counts, nil
}
