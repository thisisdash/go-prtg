package prtgapi

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func testDevicesService_List(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/table.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testParams(t, r, map[string]string{
			"filter_tags": "mytag",
		})
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(200)
		w.Write(devicesListJSON)
	})

	ctx := context.Background()
	got, err := client.Devices().List(ctx, DeviceListOptions{
		Tags: []string{"mytag"},
	})
	if err != nil {
		t.Errorf("Error while getting devices: %v", err)
	}
	if want := wantDevices; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, expected %v", got, wantDevices)
	}
}

func TestDevicesService_Duplicate(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/duplicateobject.htm", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testParams(t, r, map[string]string{
			"id":       "123",
			"targetid": "321",
			"name":     "testdevice",
			"host":     "testdevice.example.com",
		})
		w.Header().Add("Content-Type", "text/html; charset=UTF-8")
		w.Header().Add("Location", "/device.htm?id=1234")
		w.WriteHeader(302)
	})

	mux.HandleFunc("/api/table.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testParams(t, r, map[string]string{
			"filter_objid": "1234",
		})
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(200)
		w.Write(deviceListJSON)
	})

	mux.HandleFunc("/api/setobjectproperty.htm", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testParams(t, r, map[string]string{
			"id":    "1234",
			"name":  "tags",
			"value": "mytag,another-tag",
		})
		w.Header().Add("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(200)
		w.Write([]byte(`<HTML><BODY class="no-content"><B class="no-content">OK</B></BODY></HTML>`))
	})

	ctx := context.Background()
	got, err := client.Devices().Duplicate(ctx, 123, 321, "testdevice", "testdevice.example.com", []string{"mytag", "another-tag"})
	if err != nil {
		t.Errorf("Error while duplicating devices: %v", err)
	}
	if want := wantDevice; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, expected %v", got, want)
	}
}

func TestDevicesService_GetByID(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/table.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testParams(t, r, map[string]string{
			"filter_objid": "1234",
		})
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(200)
		w.Write(deviceListJSON)
	})

	ctx := context.Background()
	got, err := client.Devices().GetByID(ctx, 1234, DeviceListOptions{})
	if err != nil {
		t.Errorf("Error while getting device by ID: %v", err)
	}
	if want := wantDevice; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, expected %v", got, want)
	}
}

func TestDevicesService_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/table.json", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testParams(t, r, map[string]string{
			"filter_tags": "myapp",
		})
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(200)
		w.Write(deviceListJSON)
	})

	ctx := context.Background()
	got, err := client.Devices().GetByID(ctx, 1234, DeviceListOptions{
		Tags: []string{"myapp"},
	})
	if err != nil {
		t.Errorf("Error while getting device: %v", err)
	}
	if want := wantDevice; !reflect.DeepEqual(got, want) {
		t.Errorf("Got %v, expected %v", got, want)
	}
}

var wantDevice = &Device{
	ID:   1234,
	Name: "testdevice",
	Host: "testdevice.example.com",
}

var wantDevices = []*Device{
	wantDevice,
	&Device{
		ID:   1235,
		Name: "another",
		Host: "another.example.com",
	},
}

var deviceListJSON = []byte(`{
	"prtg-version": "19.4.53.1912",
	"treesize": 1,
	"devices": [
		{
			"objid": 1234,
			"device": "testdevice",
			"host": "testdevice.example.com"
		}
	]
}`)

var devicesListJSON = []byte(`{
	"prtg-version": "19.4.53.1912",
	"treesize": 2,
	"devices": [
		{
			"objid": 1234,
			"device": "testdevice",
			"host": "testdevice.example.com"
		},
		{
			"objid": 1235,
			"device": "another",
			"host": "another.example.com"
		}
	]
}`)
