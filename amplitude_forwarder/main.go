package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
)

func main() {}

func init() {}

// ModifierRegisterer is the symbol the plugin loader will be looking for. It must
// implement the plugin.Registerer interface
// https://github.com/luraproject/lura/blob/master/proxy/plugin/modifier.go#L71
var ModifierRegisterer = registerer("amplitude-forwarder")

type registerer string

// RegisterModifiers is the function the plugin loader will call to register the
// modifier(s) contained in the plugin using the function passed as argument.
// f will register the factoryFunc under the name and mark it as a request
// and/or response modifier.
func (r registerer) RegisterModifiers(f func(
	name string,
	factoryFunc func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r)+"-request", r.requestModifier, true, false)
	f(string(r)+"-response", r.responseModifier, false, true)
}

var errUnknownType = errors.New("unknown request type")

// RequestWrapper is an interface for passing proxy request between the krakend pipe
// and the loaded plugins
type RequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
}

// ResponseWrapper is an interface for passing proxy response between the krakend pipe
// and the loaded plugins
type ResponseWrapper interface {
	Data() map[string]interface{}
	Io() io.Reader
	IsComplete() bool
	StatusCode() int
	Headers() map[string][]string
}

func (r registerer) requestModifier(
	cfg map[string]interface{},
) func(interface{}) (interface{}, error) {
	return func(input interface{}) (interface{}, error) {
		req, ok := input.(RequestWrapper)
		if !ok {
			fmt.Println("❌ ERROR: Response is of unknown type", input)
			return nil, errUnknownType
		}

		// Transform from featureFlags to flag_keys query string
		req.Query().Set("flag_keys", req.Query().Get("featureFlags"))
		req.Query().Del("featureFlags")

		// Get deviceId from headers and set it to query string
		deviceId := req.Headers()["X-Deviceid"]
		if len(deviceId) > 0 {
			req.Query().Set("device_id", deviceId[0])
			req.Headers()["X-Deviceid"] = nil
		}

		// Add Authorization header
		// TODO: Use env variable for the API key
		req.Headers()["Authorization"] = []string{"Api-Key client-9XxS03kKfG3MMsAfafVJz8h9aJmu0qZy"}

		return input, nil
	}
}

type FeatureFlagItem struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func (r registerer) responseModifier(
	cfg map[string]interface{},
) func(interface{}) (interface{}, error) {
	return func(input interface{}) (interface{}, error) {
		resp, ok := input.(ResponseWrapper)
		if !ok {
			fmt.Println("❌ ERROR: Response is of unknown type", input)
			return nil, errUnknownType
		}

		newData := []FeatureFlagItem{}
		for key, value := range resp.Data() {
			newObj := FeatureFlagItem{
				Name:   key,
				Active: value.(map[string]interface{})["key"] == "on",
			}
			newData = append(newData, newObj)
		}

		var modified = &responseWrapper{
			data:       newData,
			io:         resp.Io(),
			complete:   resp.IsComplete(),
			statusCode: resp.StatusCode(),
			headers:    resp.Headers(),
		}

		return modified, nil
	}
}

type responseWrapper struct {
	headers    map[string][]string
	statusCode int
	complete   bool
	io         io.Reader
	data       []FeatureFlagItem
}

func (r *responseWrapper) Data() map[string]interface{} {
	// We need to return "collection" key to make it work with the krakend `json-collection` encoder
	return map[string]interface{}{"collection": r.data}
}
func (r responseWrapper) IsComplete() bool             { return r.complete }
func (r responseWrapper) StatusCode() int              { return r.statusCode }
func (r responseWrapper) Io() io.Reader                { return r.io }
func (r responseWrapper) Headers() map[string][]string { return r.headers }
