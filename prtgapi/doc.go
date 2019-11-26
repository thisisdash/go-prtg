/*
Package prtgapi provides a client for the PRTG API
that is focused on object manipulation

Sample usage

	url, err := url.Parse("https://prtg.example.com")
	if err != nil {
		log.Fatalf("Unable to parse PRTG URL: %v", err)
	}

	client := prtgapi.NewClient(
		url,
		"username",
		"passhash", // The passhash can be retrieved from the user profile page in PRTG
		"my user-agent",
		&http.Client{} // When no HTTP client is given, prtgapi will use the standard HTTP client
	)

PRTG's API is unique as it does some weird things. To make sure that the library
works the http client is configured to NOT follow redirects, this is done automatically
on the passed in http client.
*/
package prtgapi
