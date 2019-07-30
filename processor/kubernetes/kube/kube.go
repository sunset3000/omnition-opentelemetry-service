// Copyright 2019 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"fmt"
	"log"
	"time"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ExtractionRules ...
type ExtractionRules struct {
	PodName   bool
	StartTime bool
}

// Pod represents a kubernetes pod
type Pod struct {
	Name    string
	Address string
	// Attributes map[string]*tracepb.AttributeValue
	Attributes map[string]string
}

// Client is the main interface provided by this package to a kubernetes cluster
type Client struct {
	kc        *kubernetes.Clientset
	Namespace string
	Pods      map[string]Pod
	Rules     ExtractionRules
}

// New initializes a new k8s Client
func New(namespace string, rules ExtractionRules) (*Client, error) {
	k := &Client{Namespace: namespace, Rules: rules}
	err := k.start()
	return k, err
}

// PodByIP takes an IP address and returns the pod the IP address is associated with
func (c *Client) PodByIP(ip string) (Pod, bool) {
	pod, ok := c.Pods[ip]
	return pod, ok
}

func (c *Client) start() error {
	c.Pods = map[string]Pod{}
	kc, err := kubeClientset()
	if err != nil {
		return err
	}
	c.kc = kc
	return c.watch()
}

func (c *Client) getTagsForPod(pod *api_v1.Pod) map[string]string {
	tags := map[string]string{}
	if c.Rules.PodName {
		tags["k8s.pod"] = pod.Name
	}
	if c.Rules.StartTime {
		// tags["k8s.startTime"] = stringAttribute(pod.Status.StartTime.String())
		tags["k8s.startTime"] = pod.Status.StartTime.String()
	}
	return tags
}

func stringAttribute(value string) *tracepb.AttributeValue {
	return &tracepb.AttributeValue{
		Value: &tracepb.AttributeValue_StringValue{
			StringValue: &tracepb.TruncatableString{Value: value},
		},
	}
}

func (c *Client) watch() error {
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.kc.CoreV1().Pods("").List(metav1.ListOptions{})
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.kc.CoreV1().Pods("").Watch(metav1.ListOptions{})
			},
		},
		&api_v1.Pod{},
		5*time.Minute,
		cache.Indexers{},
	)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println(">> add")
			/*
				fmt.Println("========= Add")
				fmt.Println(pod, ok)
				fmt.Println("=========")
			*/
			//mObj := obj.(metav1.Object)
			//fmt.Println(mObj.GetNamespace(), "/", mObj.GetName())
			//fmt.Println(mObj.GetAnnotations())
			//fmt.Println(mObj.GetLabels())
			if pod, ok := obj.(*api_v1.Pod); ok {
				c.Pods[pod.Status.PodIP] = Pod{
					Name:       pod.Name,
					Address:    pod.Status.PodIP,
					Attributes: c.getTagsForPod(pod),
				}
				fmt.Println(ok)
				fmt.Println(pod.Status.PodIP)
			}
		},
		UpdateFunc: func(old, obj interface{}) {
			fmt.Println(">> update")
			/*
				fmt.Println("========= Update")
				mObj := obj.(metav1.Object)
				fmt.Println(mObj.GetNamespace(), "/", mObj.GetName())
				fmt.Println(mObj.GetAnnotations())
				fmt.Println(mObj.GetLabels())
				pod, ok := obj.(*api_v1.Pod)
				fmt.Println(pod, ok)
				fmt.Println("=========")
			*/
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println(">> delete")
			/*
				fmt.Println("========= Delete")
				mObj := obj.(metav1.Object)
				fmt.Println(mObj.GetNamespace(), "/", mObj.GetName())
				fmt.Println(mObj.GetAnnotations())
				fmt.Println(mObj.GetLabels())
				pod, ok := obj.(*api_v1.Pod)
				fmt.Println(pod, ok)
				fmt.Println("=========")
			*/
		},
	})
	stopCh := make(<-chan struct{})
	go informer.Run(stopCh)
	return nil
}

func (c *Client) watchInformer() error {
	factory := informers.NewSharedInformerFactory(c.kc, 0)
	informer := factory.Core().V1().Pods().Informer()
	stopper := make(chan struct{})
	// defer close(stopper)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// "k8s.io/apimachinery/pkg/apis/meta/v1" provides an Object
			// interface that allows us to get metadata easily
			mObj := obj.(metav1.Object)
			log.Printf("New Pod Added to Store: %s", mObj.GetName())
		},
	})
	informer.Run(stopper)
	return nil
}

func (c *Client) watchBasic() error {
	// filter by labels if provided
	watcher, err := c.kc.CoreV1().Pods(c.Namespace).Watch(metav1.ListOptions{})
	if err != nil {
		// wrap
		return err
	}
	go func() {
		for {
			select {
			case e := <-watcher.ResultChan():
				fmt.Println(e)
			}
		}
	}()
	return nil
}

func (c *Client) fetch() {
	k8spods, err := c.kc.CoreV1().Pods(c.Namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, p := range k8spods.Items {
		if p.Status.Phase != "Running" || p.Status.PodIP == "" {
			continue
		}
		/*
			c.Pods[p.Status.PodIP] = Pod{
				p.Name, p.Status.PodIP,
			}
		*/
	}
}
