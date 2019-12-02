package prtgsyncer

import (
	"github.com/youngcapital/go-prtg/prtgapi"
)

/*
Syncer defines the synchronization settings.
It provides a way to work with all types of input objects (ingress, struct etc)
by allowing to configure your own getter functions for each of the synchronized items

The result of the Condition func determines whether the sync should trigger

Example for a kubernetes ingress object:

	syncer := &Syncer{
		Condition: func(v interface{}) bool {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object")
			}
			// Only sync when there is a hostname
			return len(ingress.Spec.Rules) > 0 && ingress.Spec.Rules[0].Host != ""
		},

		DeviceNameGetter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object")
			}
			return ingress.Name
		},
		DeviceIdentifierGetter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object")
			}
			return string(ingress.UID)
		},
		DeviceHostnameGetter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object")
			}
			return string(ingress.Spec.Rules[0].Host)
		},
	}
*/
type Syncer struct {
	Client    *prtgapi.Client
	TagPrefix string

	TemplateDeviceID           int64
	ParentGroupID              int64
	UnpauseDeviceAfterCreation bool

	Condition              func(interface{}) bool
	DeviceNameGetter       func(interface{}) string
	DeviceIdentifierGetter func(interface{}) string
	DeviceHostnameGetter   func(interface{}) string
	SensorUpdateFields     []SensorUpdateField
}

/*
SensorUpdateField allows specifying which properties of the device sensors
should be synchronized. Similar to the PrtgSyncer, each SensorUpdateField
should define a getter function that handles getting the new property value
from the passed in object

Example for a kubernetes ingress:
	field := &SensorUpdateField{
		SensorRawType: "httpadvanced",
		FieldName: "httpurl",
		Getter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object")
			}
			u := url.URL{Scheme: "https", Host: ingress.Spec.Rules[0].Host, Path: ingress.ObjectMeta.Annotations["prometheus.io/path"]}
			return u.String()
		},
	}
*/
type SensorUpdateField struct {
	SensorRawType string
	FieldName     string
	Getter        func(interface{}) string
}

// SyncResult holds the result of the sync action
// It allows to see which settings and sensors were updated
type SyncResult struct {
	NewDevice       bool
	HostnameUpdated bool
	ConditionFailed bool

	SensorSyncResults []SensorSyncResult
}

// IsNew returns whether or not the device was newly created in PRTG during the sync
func (r *SyncResult) IsNew() bool {
	return r.NewDevice
}

// IsChanged returns whether or not the device had any updates during the sync
func (r *SyncResult) IsChanged() bool {
	if r.NewDevice || r.HostnameUpdated {
		return true
	}

	for _, sr := range r.SensorSyncResults {
		if sr.SensorUpdated {
			return true
		}
	}

	return false
}

// IsIgnored returns if the device is ignored because of a condition failure
func (r *SyncResult) IsIgnored() bool {
	return r.ConditionFailed
}

// SensorSyncResult holds the sync result of the individual sensors
type SensorSyncResult struct {
	SensorRawType string
	SensorUpdated bool
	FieldUpdated  string
}
