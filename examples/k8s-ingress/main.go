package main

import (
	"context"
	"log"
	"net/url"

	"github.com/youngcapital/go-prtg/prtgapi"
	"github.com/youngcapital/go-prtg/prtgsyncer"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Create PRTG API client
	prtgURL, _ := url.Parse("https://prtg.example.com")
	prtgclient := prtgapi.NewClient(*prtgURL, "demo", "demo", "k8s-ingress-example", nil)

	// Configure syncer
	// This expects a device with a single sensor (of type HTTP advanced)
	syncer := &prtgsyncer.Syncer{
		Client:           prtgclient,
		TemplateDeviceID: 1000,
		ParentGroupID:    900,
		TagPrefix:        "k8s-ingress",
		Condition: func(v interface{}) bool {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object, got %v", v)
			}
			return len(ingress.Spec.Rules) > 0 && ingress.Spec.Rules[0].Host != ""
		},
		DeviceNameGetter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object, got %v", v)
			}
			return ingress.Name
		},
		DeviceIdentifierGetter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object, got %v", v)
			}
			return string(ingress.UID)
		},
		DeviceHostnameGetter: func(v interface{}) string {
			ingress, ok := v.(*v1beta1.Ingress)
			if !ok {
				log.Fatalf("Didn't receive an ingress object, got %v", v)
			}
			return string(ingress.Spec.Rules[0].Host)
		},
		SensorUpdateFields: []prtgsyncer.SensorUpdateField{
			prtgsyncer.SensorUpdateField{
				SensorRawType: "httpadvanced",
				FieldName:     "httpurl",
				Getter: func(v interface{}) string {
					ingress, ok := v.(*v1beta1.Ingress)
					if !ok {
						log.Fatalf("Didn't receive an ingress object, got %v", v)
					}
					u := url.URL{Scheme: "https", Host: ingress.Spec.Rules[0].Host, Path: ingress.ObjectMeta.Annotations["prometheus.io/path"]}
					return u.String()
				},
			},
		},
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatalf("Error while getting clientconfig: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error while getting clientset: %v", err)
	}

	watcher, err := clientset.ExtensionsV1beta1().Ingresses("").Watch(v1.ListOptions{})
	if err != nil {
		log.Fatalf("Error while watching ingresses: %v", err)
	}

	ch := watcher.ResultChan()
	for event := range ch {
		ingress := event.Object.(*v1beta1.Ingress)
		result, err := syncer.Sync(context.Background(), ingress)
		if err != nil {
			log.Fatalf("Error while syncing ingress %s to PRTG: %v", ingress.Name, err)
		}
		log.Printf("Sync to PRTG results for ingress %s: %+v", ingress.Name, result)
	}
}
