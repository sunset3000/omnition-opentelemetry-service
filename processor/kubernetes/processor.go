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

package kubernetes

import (
	"context"
	"fmt"

	resourcepb "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"github.com/open-telemetry/opentelemetry-service/consumer"
	"github.com/open-telemetry/opentelemetry-service/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-service/processor"

	"github.com/Omnition/omnition-opentelemetry-service/client"
	"github.com/Omnition/omnition-opentelemetry-service/processor/kubernetes/kube"
)

type kubernetesprocessor struct {
	nextConsumer consumer.TraceConsumer
	kc           *kube.Client
	namespace    string
	rules        kube.ExtractionRules
}

// NewTraceProcessor returns a processor.TraceProcessor that adds the WithAttributeMap(attributes) to all spans
// passed to it.
func NewTraceProcessor(nextConsumer consumer.TraceConsumer, options ...Option) (processor.TraceProcessor, error) {
	kp := &kubernetesprocessor{nextConsumer: nextConsumer}
	for _, opt := range options {
		if err := opt(kp); err != nil {
			return nil, err
		}
	}
	kc, err := kube.New(kp.namespace, kp.rules)
	if err != nil {
		return nil, err
	}
	kp.kc = kc
	// fmt.Println(kp.kc.Pods)
	return kp, nil
}

func (kp *kubernetesprocessor) ConsumeTraceData(ctx context.Context, td consumerdata.TraceData) error {
	c, ok := client.FromContext(ctx)
	if !ok {
		return kp.nextConsumer.ConsumeTraceData(ctx, td)
	}
	attrs := kp.getAttributesForClient(c)
	if len(attrs) == 0 {
		return kp.nextConsumer.ConsumeTraceData(ctx, td)
	}

	// should we even be adding resource if it is not suplied by the exporter?
	if td.Resource == nil {
		td.Resource = &resourcepb.Resource{}
	}

	if td.Resource != nil {
		if td.Resource.Labels == nil {
			td.Resource.Labels = map[string]string{}
		}
		for k, v := range attrs {
			td.Resource.Labels[k] = v
		}
	}

	// TODO: should add to spans that have a resource not the same as the batch
	for _, span := range td.Spans {
		if span == nil {
			// We will not create nil spans with just attributes on them
			continue
		}
		if span.Resource != nil {
			// Add tags to span Resource
			/*
				if span.Attributes == nil {
					span.Attributes = &tracepb.Span_Attributes{}
				}
				// Create a new map if one does not exist. Could re-use passed in map, but
				// feels too unsafe.
				if span.Attributes.AttributeMap == nil {
					span.Attributes.AttributeMap = map[string]*tracepb.AttributeValue{}
				}
				// Add k8s resource tags
				for key, value := range attrs {
					if _, exists := span.Attributes.AttributeMap[key]; !exists {
						fmt.Println("adding tag :", value)
						span.Attributes.AttributeMap[key] = value
					}
				}
			*/
		}
	}
	return kp.nextConsumer.ConsumeTraceData(ctx, td)
}

func (kp *kubernetesprocessor) getAttributesForClient(c *client.Client) map[string]string {
	pod, ok := kp.kc.PodByIP(c.IP)
	if !ok {
		fmt.Println("could not find pod for ip: ", c.IP)
		return nil
	}
	return pod.Attributes
}
