package prtgapi

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// DevicesService handles communication with the device related methods of the PRTG API
type DevicesService service

type deviceList struct {
	Items []*Device `json:"devices"`
}

// Device represents a PRTG device
type Device struct {
	ID   int64  `json:"objid"`
	Name string `json:"device"`
	Host string
}

// DeviceListOptions can be used to filter devices when calling List or Get*
//
// Currently it is possible to filter on
// * ID (note, most of the times this refers to the ID of the parent)
// * Tags
// * Other filters, like
//	map[string]string{
//		"objid": "12345"
//	}
type DeviceListOptions struct {
	ID     int64
	Tags   []string
	Filter map[string]string
}

const (
	deviceListPath              = "/api/table.json"
	devicePausePath             = "/api/pause.htm"
	duplicateDevicePath         = "/api/duplicateobject.htm"
	setDeviceUpdatePropertyPath = "/api/setobjectproperty.htm"
)

// NewDevicesService returns a new DevicesService for a given client
func NewDevicesService(client *Client) *DevicesService {
	return &DevicesService{
		client: client,
	}
}

// Duplicate duplicates the device identified by templateDeviceID into a group
// identified by parentGroupID.
//
// Tags can be left empty if you don't want your new device to be tagged.
func (d *DevicesService) Duplicate(ctx context.Context, templateDeviceID int64, parentGroupID int64, name string, hostname string, tags []string) (*Device, error) {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(templateDeviceID, 10))
	v.Set("targetid", strconv.FormatInt(parentGroupID, 10))
	v.Set("name", name)
	v.Set("host", hostname)

	res := &redirectResponse{}
	err := d.client.do(ctx, duplicateDevicePath, v, res)
	if err != nil {
		return nil, err
	}

	// Get the deviceID from the redirect response
	newDeviceURL, err := url.Parse(res.Location)
	if err != nil {
		return nil, err
	}

	newDeviceID, err := strconv.ParseInt(newDeviceURL.Query().Get("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	newDevice, err := d.GetByID(ctx, newDeviceID, DeviceListOptions{})
	if err != nil {
		return nil, err
	}

	if len(tags) > 0 {
		err = d.UpdateProperty(ctx, newDevice.ID, "tags", strings.Join(tags, ","))
		if err != nil {
			return nil, err
		}
	}

	return newDevice, nil
}

// List returns a list of devices that match the given options
func (d *DevicesService) List(ctx context.Context, options DeviceListOptions) ([]*Device, error) {
	v := url.Values{}
	v.Set("content", "devices")
	v.Set("columns", "objid,device,host")
	if options.ID != 0 {
		v.Set("id", strconv.FormatInt(options.ID, 10))
	}
	if len(options.Tags) > 0 {
		for _, tag := range options.Tags {
			v.Set("filter_tags", tag)
		}
	}
	if len(options.Filter) > 0 {
		for key, value := range options.Filter {
			v.Set("filter_"+key, value)
		}
	}

	deviceList := &deviceList{}
	err := d.client.do(ctx, deviceListPath, v, deviceList)
	if err != nil {
		return nil, err
	}

	return deviceList.Items, nil
}

// Get returns a single device if there is only one device matching the given options
func (d *DevicesService) Get(ctx context.Context, options DeviceListOptions) (*Device, error) {
	return d.get(ctx, options)
}

// GetByID returns a single device identified by the id and given options.
func (d *DevicesService) GetByID(ctx context.Context, id int64, options DeviceListOptions) (*Device, error) {
	if options.Filter == nil {
		options.Filter = map[string]string{
			"objid": strconv.FormatInt(id, 10),
		}
	} else {
		options.Filter["objid"] = strconv.FormatInt(id, 10)
	}

	return d.get(ctx, options)
}

// UpdateProperty updates a property on a device
func (d *DevicesService) UpdateProperty(ctx context.Context, id int64, name string, value string) error {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("name", name)
	v.Set("value", value)
	return d.client.do(ctx, setSensorObjectPropertyPath, v, nil)
}

// Pause pauses the device indefinitely
func (d *DevicesService) Pause(ctx context.Context, id int64, message string) error {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("action", "0")
	v.Set("message", message)
	return d.client.do(ctx, devicePausePath, v, nil)
}

// Unpause unpauses the device
func (d *DevicesService) Unpause(ctx context.Context, id int64) error {
	v := url.Values{}
	v.Set("id", strconv.FormatInt(id, 10))
	v.Set("action", "1")
	return d.client.do(ctx, devicePausePath, v, nil)
}

func (d *DevicesService) get(ctx context.Context, options DeviceListOptions) (*Device, error) {
	devices, err := d.List(ctx, options)
	if err != nil {
		return nil, err
	}

	switch len(devices) {
	case 0:
		return nil, nil
	case 1:
		device := devices[0]
		return device, nil
	default:
		return nil, fmt.Errorf("More than one device matched the query")
	}
}
