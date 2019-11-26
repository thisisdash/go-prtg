package prtgapi

import (
	"context"
	"net/url"
	"strconv"
)

// SensorsService handles communication with the sensor related methods of the PRTG API
type SensorsService service

type sensorList struct {
	Items []*Sensor `json:"sensors"`
}

// Sensor represents a PRTG sensor
type Sensor struct {
	ID      int64 `json:"objid"`
	Name    string
	Type    string
	RawType string `json:"type_raw"`
}

// SensorListOptions can be used to filter sensors when calling List
//
// Currently it is possible to filter on
// * ID (this refers to the ID of the device)
// * Tags
type SensorListOptions struct {
	ID   int64
	Tags []string
}

const (
	sensorListPath              = "/api/table.json"
	sensorPausePath             = "/api/pause.htm"
	getSensorObjectPropertyPath = "/api/getobjectproperty.htm"
	setSensorObjectPropertyPath = "/api/setobjectproperty.htm"
)

// NewSensorsService returns a new SensorsService for a client
func NewSensorsService(client *Client) *SensorsService {
	s := &SensorsService{
		client: client,
	}
	return s
}

// List returns a list of sensor objects that match the given options
func (s *SensorsService) List(ctx context.Context, options SensorListOptions) ([]*Sensor, error) {
	v := url.Values{}
	v.Set("content", "sensors")
	v.Set("columns", "objid,type,type_raw,name")
	if options.ID != 0 {
		v.Set("id", strconv.FormatInt(options.ID, 10))
	}
	if len(options.Tags) > 0 {
		for _, tag := range options.Tags {
			v.Set("filter_tags", tag)
		}
	}

	sensorList := &sensorList{}
	err := s.client.do(ctx, sensorListPath, v, sensorList)
	if err != nil {
		return nil, err
	}

	return sensorList.Items, nil
}

// GetProperty returns the current value of a property
func (s *SensorsService) GetProperty(ctx context.Context, id int64, name string) (string, error) {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("name", name)

	var propertyResult struct {
		Value string `xml:"result"`
	}

	err := s.client.do(ctx, getSensorObjectPropertyPath, v, &propertyResult)
	if err != nil {
		return "", err
	}

	return propertyResult.Value, nil
}

// UpdateProperty sets a new value for a property
func (s *SensorsService) UpdateProperty(ctx context.Context, id int64, name string, value string) error {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("name", name)
	v.Set("value", value)
	return s.client.do(ctx, setSensorObjectPropertyPath, v, nil)
}

// Pause pauses a sensor indefinitely
func (s *SensorsService) Pause(ctx context.Context, id int64, message string) error {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("action", "0")
	v.Set("message", message)
	return s.client.do(ctx, sensorPausePath, v, nil)
}

// Unpause unpauses a sensor
func (s *SensorsService) Unpause(ctx context.Context, id int64) error {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("action", "1")
	return s.client.do(ctx, sensorPausePath, v, nil)
}
