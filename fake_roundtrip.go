package fake_round_trip

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

const REDIRECTED_LOCATION = "/new-location/"

func NewFakeClient() *FakeRoundTripClient {
	fakeRoundTripAgent :=  NewFakeRoundTripAgent()
	fakeClient := &FakeRoundTripClient{
		fakeRoundTripAgent: fakeRoundTripAgent,
	}

	fakeClient.Transport = fakeRoundTripAgent
	return fakeClient
}

type FakeRoundTripClient struct {
	http.Client
	fakeRoundTripAgent *FakeRoundTripAgent
}

func (f FakeRoundTripClient) AddTrip(method, url string, statusCode int, document string) *FakeRoundTrip {
	fr := &FakeRoundTrip {
		method:     method,
		url:        url,
		document:   document,
		statusCode: statusCode,
		header: &http.Header{},
	}

	f.fakeRoundTripAgent.add(url, method, *fr)
	return fr
}


func (f FakeRoundTripClient) PlanGet(url string, statusCode int, document string) *FakeRoundTrip {
	return f.AddTrip("GET", url, statusCode, document)
}

func (f FakeRoundTripClient) PlanPost(url string, statusCode int, document string) *FakeRoundTrip {
	return f.AddTrip("POST", url, statusCode, document)
}

func (f FakeRoundTripClient) PlanPut(url string, statusCode int, document string) *FakeRoundTrip {
	return f.AddTrip("PUT", url, statusCode, document)
}

func (f FakeRoundTripClient) PlanDelete(url string, statusCode int, document string) *FakeRoundTrip {
	return f.AddTrip("DELETE", url, statusCode, document)
}

func (f FakeRoundTrip) SetStatusCode(code int) *FakeRoundTrip {
	f.statusCode = code

	locationRequired := code == 302 || code == 201 || code == 202

	if locationRequired {
		setDefaultLocationHeader(&f)
	}

	return &f
}

func (f FakeRoundTrip) SetResponseHeader(key string, value string) *FakeRoundTrip {
	f.header.Set(key, value)
	return &f
}

func (f FakeRoundTrip) SetURL(url string) *FakeRoundTrip {
	f.url = url
	return &f
}

func NewFakeRoundTripAgent() *FakeRoundTripAgent {
	return &FakeRoundTripAgent{ roundTrips: make(map[string]http.RoundTripper) }
}

type FakeRoundTripAgent struct {
	roundTrips map[string]http.RoundTripper
}

func (f FakeRoundTripAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	if roundTrip := f.roundTrips[f.getKey(*r)]; roundTrip != nil {
		return roundTrip.RoundTrip(r)
	}

	return FourOFour(), nil
}

func (f FakeRoundTripAgent) add(url, method string, roundTrip FakeRoundTrip) {
	roundTrip.fakeRoundTripAgent = &f

	key :=	f.makeKey(url, method)
	f.roundTrips[key] = roundTrip
}

func (f FakeRoundTripAgent) makeKey(url, method string) string {
	return url + ":" + method
}

func (f FakeRoundTripAgent) getKey(r http.Request) string {
	return r.URL.String() + ":" + r.Method
}

type FakeRoundTrip struct {
	statusCode int
	method     string
	url        string
	document   string
	header 	*http.Header
	fakeRoundTripAgent *FakeRoundTripAgent
}

func (f FakeRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) {
	var statusCode int = f.statusCode
	expectedURL, _ := url.Parse(f.url)

	if !urlMatches(*r.URL, *expectedURL) {
		statusCode = 404
		return FourOFour(), nil
	}

//	if (statusCode == 302) {
//		redirectedURL := r.URL.Scheme + "://" + r.URL.Host + REDIRECTED_LOCATION
//		f.fakeRoundTripAgent.add(redirectedURL *****************)
//		fmt.Println("URL reset to ", f.url, " for next req")
//		fmt.Println("\n\n\n")
//	}

	expectedURL, _ = url.Parse(f.url)

	resp := &http.Response {
		Body:       NewFakeReadCloser(f.document),
		StatusCode: statusCode,
		Header: *f.header,
	}

	return resp, nil
}

func urlMatches(actual url.URL, expected url.URL) bool {
	match := (actual.Scheme == expected.Scheme) && (actual.Host == expected.Host) && (actual.Path == expected.Path) && (actual.RawQuery == expected.RawQuery)
	if match {
		return true
	}

	return false
}

//per handle location headers: http://en.wikipedia.org/wiki/HTTP_location
func setDefaultLocationHeader(f *FakeRoundTrip) {
	f.SetResponseHeader("Location", REDIRECTED_LOCATION)
}

func FourOFour() *http.Response {
	resp := &http.Response {
		Body:    NewFakeReadCloser("Unknown"),
		StatusCode: 404,
	}

	return resp
}

func NewFakeReadCloser(body string) *FakeReadCloser {
	fr := &FakeReadCloser{
		Reader: strings.NewReader(body),
	}

	return fr
}

type FakeReadCloser struct {
	io.Reader
}

func (f FakeReadCloser) Close() error {
	return nil
}
