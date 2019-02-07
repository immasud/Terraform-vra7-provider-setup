package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	logging "github.com/op/go-logging"
	"github.com/vmware/terraform-provider-vra7/utils"
)

var (
	log       = logging.MustGetLogger(utils.LoggerID)
	apiClient APIClient
)

const (
	// Common header key
	ContentTypeHeader   = "Content-Type"
	AuthorizationHeader = "Authorization"
	ConnectionHeader    = "Connection"
	AcceptHeader        = "Accept"

	// Common header values
	AppJSON         = "application/json"
	CloseConnection = "close"

	// Http Methods
	GET    = "GET"
	POST   = "POST"
	PATCH  = "PATCH"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type APIClient struct {
	Username    string
	Password    string
	BaseURL     string
	Tenant      string
	Insecure    bool
	BearerToken string
	Client      *http.Client
}

func (ar *APIRequest) AddHeader(key, val string) {
	if ar.Headers == nil {
		ar.Headers = make(map[string]string)
	}
	ar.Headers[key] = val
}

func (ar *APIRequest) ContentType() string {
	if ar.Headers == nil {
		return ""
	}

	contentType, ok := ar.Headers[ContentTypeHeader]
	if !ok {
		return ""
	}
	return contentType
}

// CopyHeadersTo Add headers to request object
func (ar *APIRequest) CopyHeadersTo(req *http.Request) {
	for key, val := range ar.Headers {
		req.Header.Add(key, val)
	}
}

// Error Implement Go error interface for ApiError
func (e APIError) Error() string {
	return fmt.Sprintf("Error: { HTTP status: '%v', message: '%v'}",
		e.HttpStatusCode, e.Message)
}

// GetAPIError reads an error out of the HTTP response, or does nothing if
// no error occured.
func GetAPIError(respBody []byte, statusCode int) error {
	apiError := APIError{}
	unmarshalErr := json.Unmarshal(respBody, &apiError)
	if unmarshalErr != nil {
		// Do not return this error just log it.
		log.Error("Error is %v ", unmarshalErr)
		apiError.Message = string(respBody)
	}

	apiError.HttpStatusCode = statusCode
	return apiError
}

// FromAPIRequestToHTTPRequest converts API request object to http request
func FromAPIRequestToHTTPRequest(apiReq *APIRequest) (*http.Request, error) {
	req, err := http.NewRequest(apiReq.Method, apiReq.URL, apiReq.Body)
	if err != nil {
		return nil, err
	}

	if apiReq.ContentType() == "" {
		apiReq.AddHeader(ContentTypeHeader, AppJSON)
	}
	apiReq.CopyHeadersTo(req)
	return req, nil
}

// FromHTTPRespToAPIResp converts Http response to API response
func FromHTTPRespToAPIResp(resp *http.Response) (*APIResponse, error) {
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, GetAPIError(respBody, resp.StatusCode)
	}

	apiResp := &APIResponse{}
	apiResp.Body = respBody
	apiResp.Headers = resp.Header
	apiResp.Status = resp.Status
	apiResp.StatusCode = resp.StatusCode
	return apiResp, nil
}

// BuildEncodedURL build the url by adding the base url and headers, etc
func BuildEncodedURL(relativePath string, queryParameters map[string]string) string {
	//Todo it might be better to swith to Viper to load all the config at once
	serverURL := apiClient.BaseURL

	var queryURL *url.URL
	queryURL, err := url.Parse(serverURL)
	if err != nil {
		log.Error("Error %v ", err)
	}

	queryURL.Path += relativePath

	parameters := url.Values{}
	if queryParameters != nil {
		for key, value := range queryParameters {
			parameters.Add(key, value)
		}
	} else {
		return queryURL.String()
	}
	queryURL.RawQuery = parameters.Encode()
	return queryURL.String()
}