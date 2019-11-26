package prtgsyncer

import (
	"context"
	"fmt"

	"github.com/youngcapital/go-prtg/prtgapi"
)

// Sync synchronizes a device to PRTG given an object
func (s *Syncer) Sync(ctx context.Context, v interface{}) (*SyncResult, error) {
	result := &SyncResult{}

	if !s.Condition(v) {
		result.ConditionFailed = true
		return result, nil
	}

	// Find the device in PRTG
	deviceIdentifier := s.DeviceIdentifierGetter(v)
	identifyingTags := []string{
		s.TagPrefix,
		s.TagPrefix + "-id-" + deviceIdentifier,
	}

	device, err := s.Client.Devices().Get(ctx, prtgapi.DeviceListOptions{
		Tags: identifyingTags,
	})
	if err != nil {
		return nil, err
	}

	if device == nil {
		deviceName := s.DeviceNameGetter(v)
		hostname := s.DeviceHostnameGetter(v)
		device, err = s.Client.Devices().Duplicate(ctx, s.TemplateDeviceID, s.ParentGroupID, deviceName, hostname, identifyingTags)
		if err != nil {
			return nil, err
		}
		result.NewDevice = true
		if device == nil {
			return nil, fmt.Errorf("Created device %s, but no device was found when getting it from the API", deviceName)
		}
	}

	// Compare hostname and update if necessary
	if newHostname := s.DeviceHostnameGetter(v); newHostname != device.Host {
		err = s.Client.Devices().UpdateProperty(ctx, device.ID, "Host", newHostname)
		if err != nil {
			return nil, err
		}
		result.HostnameUpdated = true
	}

	if len(s.SensorUpdateFields) == 0 {
		return result, nil
	}

	sensors, err := s.Client.Sensors().List(ctx, prtgapi.SensorListOptions{
		ID: device.ID,
	})
	if err != nil {
		return result, err
	}

	// Compare sensor attributes
	for _, suf := range s.SensorUpdateFields {
		for _, sensor := range sensors {
			if sensor.RawType != suf.SensorRawType {
				continue
			}

			propValue, err := s.Client.Sensors().GetProperty(ctx, sensor.ID, suf.FieldName)
			if err != nil {
				return result, err
			}

			sufValue := suf.Getter(v)
			if propValue == sufValue {
				result.SensorSyncResults = append(result.SensorSyncResults, SensorSyncResult{SensorRawType: suf.SensorRawType, SensorUpdated: false, FieldUpdated: suf.FieldName})
			} else {
				err = s.Client.Sensors().UpdateProperty(ctx, sensor.ID, suf.FieldName, sufValue)
				if err != nil {
					return result, err
				}
				result.SensorSyncResults = append(result.SensorSyncResults, SensorSyncResult{SensorRawType: suf.SensorRawType, SensorUpdated: true, FieldUpdated: suf.FieldName})
			}
		}
	}

	return result, nil
}
