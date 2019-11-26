# Go PRTG api tooling

This repository holds 2 libraries that can be used for connecting to the PRTG api.

We use this to read our kubernetes ingresses and automatically create/update devices
in PRTG based on the information read from kubernetes.

## prtgapi

prtgapi is a wrapper around the PRTG api focused on manipulation of devices/sensors.

Usage example:

```
u := url.Parse("https://prtg.example.com")
httpClient := &http.Client{} // Optional, you can also pass in nil to use the default http client
prtgclient := prtgapi.NewClient(u, "myuser", "mypasshash", "my-user-agent", httpClient)

devices, err := prtgclient.Devices().List()
if err != nil {
  log.Fatalf("Got error while fetching devices: %v", err)
}
log.Printf("Got devices: %+v", devices)
```

## prtgsyncer

prtgsyncer is a library that can handle the synchronization of devices/sensors in PRTG based
on an input object (e.g. a kubernetes ingress of some other kind of object)

Please refer to the documentation on the Syncer object for usage information.
