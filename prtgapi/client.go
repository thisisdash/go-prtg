package prtgapi

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

// Client holds the PRTG api client
// Use NewClient to create a new client
type Client struct {
	URL      url.URL
	Username string
	Passhash string

	UserAgent  string
	HTTPClient *http.Client

	devicesService *DevicesService
	sensorsService *SensorsService
}

type redirectResponse struct {
	Location string
}

// NewClient creates a new PRTG api client
func NewClient(url url.URL, username string, passhash string, userAgent string, httpClient *http.Client) *Client {
	client := &Client{
		URL:        url,
		Username:   username,
		Passhash:   passhash,
		UserAgent:  userAgent,
		HTTPClient: httpClient,
	}

	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{}
	}

	// Make sure the PRTG client doesn't follow redirects.
	// When creating a new device the redirect is used to give back the new device ID in the location header
	client.HTTPClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	client.devicesService = NewDevicesService(client)
	client.sensorsService = NewSensorsService(client)

	return client
}

// Devices provides access to the API actions that apply to devices
func (client *Client) Devices() *DevicesService {
	return client.devicesService
}

// Sensors provides access to the API actions that apply to devices
func (client *Client) Sensors() *SensorsService {
	return client.sensorsService
}

func (client *Client) do(ctx context.Context, path string, values url.Values, v interface{}) error {
	values.Set("username", client.Username)
	values.Set("passhash", client.Passhash)

	url := client.URL
	url.Path = path
	url.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return err
	}

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Error while authenticating to PRTG")
	}

	if res.StatusCode == http.StatusFound {
		redirect, ok := v.(*redirectResponse)
		if !ok {
			return fmt.Errorf("Got a 302 redirect response from PRTG but no redirectResponse object was passed in")
		}
		redirect.Location = res.Header.Get("Location")
		return nil
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Got a non-200 response from PRTG. Status %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	isJSON, err := isJSON(res.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	if isJSON {
		err = json.Unmarshal(body, &v)
		if err != nil {
			return err
		}
		return nil
	}

	err = xml.Unmarshal(body, &v)
	if err != nil {
		return err
	}

	return nil
}

func isJSON(contentType string) (bool, error) {
	for _, c := range strings.Split(contentType, ",") {
		mediatype, _, err := mime.ParseMediaType(c)
		if err != nil {
			return false, err
		}
		if mediatype == "application/json" {
			return true, nil
		}
	}
	return false, nil
}
