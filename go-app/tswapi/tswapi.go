package tswapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
	"tsw_controller_app/logger"
)

type TSWAPIConfig struct {
	BaseURL    string `example:"http://localhost:31270"`
	CommAPIKey string
}

type ITSWAPI interface {
	ListCurrentDrivableActor() (TSWAPI_ListResponse, error)
	GetCurrentDrivableActorObjectClass() (string, error)
	DeleteSubscription(id int) error
	GetSubscription(id int) (map[string]any, error)
	SetInteracting(control string, value float64) error
	SetInputValue(control string, value float64) error
	GetInputValue(control string) (float64, error)
	GetIsButton(control string) (bool, error)
	GetActiveCab() (TSWAPIActiveCab, error)
	CreateCurrentDrivableActorSubscription(id int) error
	GetCurrentDrivableActorSubscription(id int) (TSWAPI_GetCurrentDrivableActorSubscriptionResponse, error)
	LoadAPIKey(path string) error
	CanConnect() bool
	Enabled() bool
}

type TSWAPI struct {
	transport  *http.Transport
	client     *http.Client
	canConnect bool
	Config     TSWAPIConfig
}

type TSWAPIActiveCab struct {
	Front bool
	Back  bool
}

var _ ITSWAPI = &TSWAPI{}

var ErrMissingCommAPIKey = errors.New("missing CommAPIKey")
var ErrNonSuccessStatusCode = errors.New("non-successfull status code returned from API")

func (c *TSWAPI) parseApiResponse(r io.ReadCloser) (map[string]any, error) {
	var data map[string]any
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}

	if error_code, has_error_code := data["errorCode"]; has_error_code {
		return nil, fmt.Errorf("%s: %s", error_code.(string), data["errorMessage"].(string))
	}

	result, has_result_key := data["Result"]
	if has_result_key && result.(string) == "Error" {
		return nil, fmt.Errorf("%s: %s", result, data["Message"].(string))
	}

	return data, nil
}

func (c *TSWAPI) executeTswApiRequest(req *http.Request) (map[string]any, error) {
	if c.Config.CommAPIKey == "" {
		return nil, ErrMissingCommAPIKey
	}

	try_count := 0
	for {
		req.Header.Add("DTGCommKey", c.Config.CommAPIKey)
		resp, err := c.client.Do(req)
		c.canConnect = err == nil || (resp != nil && resp.StatusCode != 401 && resp.StatusCode != 403)

		/* an error here generally always means some kind of connection error which we could retry */
		if err != nil {
			logger.Logger.Error("[TSWAPI::executeTswApiRequest] API request failed", "error", err)

			var op_err *net.OpError
			if errors.As(err, &op_err) && op_err.Err != nil {
				error_str := op_err.Err.Error()
				if error_str == "connect: connection refused" {
					return nil, fmt.Errorf("could not connect to API: %w", err)
				}
			}

			if try_count < 3 {
				try_count++
				continue
			}

			return nil, fmt.Errorf("api error: %w", err)
		}

		if resp.StatusCode >= 300 {
			logger.Logger.Error("[TSWAPI::executeTswApiRequest] API request failed", "status", resp.StatusCode)
			return nil, ErrNonSuccessStatusCode
		}

		defer resp.Body.Close()
		return c.parseApiResponse(resp.Body)
	}
}

func (c *TSWAPI) ListCurrentDrivableActor() (TSWAPI_ListResponse, error) {
	list_path := "/list/CurrentDrivableActor"
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, list_path)
	list_req, _ := http.NewRequest("GET", req_url, nil)

	response, err := c.executeTswApiRequest(list_req)
	if err != nil {
		return TSWAPI_ListResponse{}, err
	}

	raw_nodes := response["Nodes"].([]any)
	nodes := []TSWAPI_ListResponse_Node{}
	for _, node := range raw_nodes {
		nodes = append(nodes, TSWAPI_ListResponse_Node{
			Name: node.(map[string]any)["Name"].(string),
		})
	}

	return TSWAPI_ListResponse{
		Nodes: nodes,
	}, nil
}

func (c *TSWAPI) GetCurrentDrivableActorObjectClass() (string, error) {
	get_path := "/get/CurrentDrivableActor.ObjectClass"
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, get_path)
	set_req, _ := http.NewRequest("GET", req_url, nil)
	data, err := c.executeTswApiRequest(set_req)
	if err != nil {
		return "", err
	}
	values := data["Values"].(map[string]any)
	return values["ObjectClass"].(string), nil
}

func (c *TSWAPI) DeleteSubscription(id int) error {
	sub_path := fmt.Sprintf("/subscription?Subscription=%d", id)
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, sub_path)
	req, _ := http.NewRequest("DELETE", req_url, nil)
	if _, err := c.executeTswApiRequest(req); err != nil {
		return err
	}
	return nil
}

func (c *TSWAPI) GetSubscription(id int) (map[string]any, error) {
	sub_path := fmt.Sprintf("/subscription?Subscription=%d", id)
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, sub_path)
	req, _ := http.NewRequest("GET", req_url, nil)
	data, err := c.executeTswApiRequest(req)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *TSWAPI) SetInteracting(control string, value float64) error {
	set_path := fmt.Sprintf("/set/CurrentDrivableActor/%s.Interacting?Value=%f", url.PathEscape(control), value)
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, set_path)
	set_req, _ := http.NewRequest("PATCH", req_url, nil)
	if _, err := c.executeTswApiRequest(set_req); err != nil {
		return err
	}
	return nil
}

func (c *TSWAPI) SetInputValue(control string, value float64) error {
	set_path := fmt.Sprintf("/set/CurrentDrivableActor/%s.InputValue?Value=%f", url.PathEscape(control), value)
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, set_path)
	set_req, _ := http.NewRequest("PATCH", req_url, nil)
	if _, err := c.executeTswApiRequest(set_req); err != nil {
		return err
	}
	return nil
}

func (c *TSWAPI) GetInputValue(control string) (float64, error) {
	set_path := fmt.Sprintf("/get/CurrentDrivableActor/%s.InputValue", url.PathEscape(control))
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, set_path)
	set_req, _ := http.NewRequest("GET", req_url, nil)
	data, err := c.executeTswApiRequest(set_req)
	if err != nil {
		return 0, err
	}
	values := data["Values"].(map[string]any)
	return values["InputValue"].(float64), nil
}

func (c *TSWAPI) GetIsButton(control string) (bool, error) {
	set_path := fmt.Sprintf("/get/CurrentDrivableActor/%s.ObjectClass", url.PathEscape(control))
	req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, set_path)
	set_req, _ := http.NewRequest("GET", req_url, nil)
	data, err := c.executeTswApiRequest(set_req)
	if err != nil {
		return false, err
	}
	values := data["Values"].(map[string]any)
	return values["ObjectClass"].(string) == "PushButtonComponent", nil
}

func (c *TSWAPI) GetActiveCab() (TSWAPIActiveCab, error) {
	req_url := fmt.Sprintf("%s/get/CurrentDrivableActor.Function.IS_GetActiveCab", c.Config.BaseURL)
	set_req, _ := http.NewRequest("GET", req_url, nil)
	data, err := c.executeTswApiRequest(set_req)
	if err != nil {
		return TSWAPIActiveCab{}, err
	}

	active_cab := TSWAPIActiveCab{}
	values := data["Values"].(map[string]any)

	if bFront, hasValue := values["bFront"]; hasValue {
		if boolVal, ok := bFront.(bool); ok {
			active_cab.Front = boolVal
		}
	}

	if bBack, hasValue := values["bBack"]; hasValue {
		if boolVal, ok := bBack.(bool); ok {
			active_cab.Back = boolVal
		}
	}

	return active_cab, nil
}

func (c *TSWAPI) CreateCurrentDrivableActorSubscription(id int) error {
	actor_list, err := c.ListCurrentDrivableActor()
	if err != nil {
		return err
	}

	subscribe_names := []string{
		"CurrentDrivableActor.ObjectClass",
	}
	for _, node := range actor_list.Nodes {
		if _, err := c.GetInputValue(node.Name); err == nil {
			subscribe_names = append(subscribe_names, fmt.Sprintf("CurrentDrivableActor/%s.Property.InputIdentifier", url.PathEscape(node.Name)))
			subscribe_names = append(subscribe_names, fmt.Sprintf("CurrentDrivableActor/%s.InputValue", url.PathEscape(node.Name)))
			subscribe_names = append(subscribe_names, fmt.Sprintf("CurrentDrivableActor/%s.Function.GetNormalisedInputValue", url.PathEscape(node.Name)))
		}
	}
	for _, subscribe_name := range subscribe_names {
		post_path := fmt.Sprintf("/subscription/%s?Subscription=%d", subscribe_name, id)
		req_url := fmt.Sprintf("%s%s", c.Config.BaseURL, post_path)
		set_req, _ := http.NewRequest("POST", req_url, nil)
		_, err := c.executeTswApiRequest(set_req)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *TSWAPI) GetCurrentDrivableActorSubscription(id int) (TSWAPI_GetCurrentDrivableActorSubscriptionResponse, error) {
	data, err := c.GetSubscription(id)
	if err != nil {
		return TSWAPI_GetCurrentDrivableActorSubscriptionResponse{}, err
	}

	response := TSWAPI_GetCurrentDrivableActorSubscriptionResponse{
		Controls: map[string]TSWAPI_GetCurrentDrivableActorSubscriptionResponse_Control{},
	}
	control_path_rx := regexp.MustCompile(`^CurrentDrivableActor\/([^.]+)\.(.+)$`)
	entries := data["Entries"].([]any)
	for _, raw_entry := range entries {
		entry := raw_entry.(map[string]any)
		if !entry["NodeValid"].(bool) {
			/* this could happen for various reasons; such as the loco being deleted */
			continue
		}

		path := entry["Path"].(string)
		values_raw := entry["Values"]
		if values_raw == nil {
			continue
		}

		values := values_raw.(map[string]any)
		if path == "CurrentDrivableActor.ObjectClass" {
			response.ObjectClass = values["ObjectClass"].(string)
		} else {
			rx_result := control_path_rx.FindStringSubmatch(path)
			if rx_result != nil {
				control_name := rx_result[1]
				control_value_type := rx_result[2] /* InputValue, Property.InputIdentifier or Function.GetNormalisedInputValue */
				existing_entry := response.Controls[control_name]
				existing_entry.PropertyName = control_name
				switch control_value_type {
				case "InputValue":
					existing_entry.CurrentValue = values["InputValue"].(float64)
				case "Property.InputIdentifier":
					existing_entry.Identifier = values["identifier"].(string)
				case "Function.GetNormalisedInputValue":
					existing_entry.CurrentNormalizedValue = values["ReturnValue"].(float64)
				}
				response.Controls[control_name] = existing_entry
			}
		}
	}

	return response, nil
}

func (c *TSWAPI) LoadAPIKey(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}

	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	key_bytes, err := io.ReadAll(fh)
	if err != nil {
		return err
	}

	c.Config.CommAPIKey = string(key_bytes)
	return nil
}

func (c *TSWAPI) CanConnect() bool {
	return c.canConnect
}

func (c *TSWAPI) Enabled() bool {
	return c.Config.CommAPIKey != ""
}

func NewTSWAPI(config TSWAPIConfig) *TSWAPI {
	transport := &http.Transport{
		DisableKeepAlives: true,
	}
	conn := TSWAPI{
		transport:  transport,
		client:     &http.Client{Transport: transport, Timeout: 10 * time.Second},
		canConnect: false,
		Config:     config,
	}
	return &conn
}
