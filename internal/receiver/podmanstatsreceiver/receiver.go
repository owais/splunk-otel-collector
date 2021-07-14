// Copyright 2020 OpenTelemetry Authors
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

package podmanstatsreceiver

import (
	"context"
	"fmt"
	"net/url"
	"time"
	// "sync"

	// "github.com/containers/podman/v3"
	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/obsreport"
	// "go.opentelemetry.io/collector/translator/internaldata"
	"go.uber.org/zap"
)

var _ component.MetricsReceiver = (*Receiver)(nil)

type Receiver struct {
	config       *Config
	logger       *zap.Logger
	nextConsumer consumer.Metrics
	// client            *dockerClient
	podmanCtx         context.Context
	obsCtx            context.Context
	runnerCtx         context.Context
	runnerCancel      context.CancelFunc
	successfullySetup bool
	transport         string
	obsrecv           *obsreport.Receiver
}

func NewReceiver(
	_ context.Context,
	logger *zap.Logger,
	config *Config,
	nextConsumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not determine receiver transport: %w", err)
	}

	receiver := Receiver{
		config:       config,
		nextConsumer: nextConsumer,
		logger:       logger,
		transport:    parsed.Scheme,
		obsrecv:      obsreport.NewReceiver(obsreport.ReceiverSettings{ReceiverID: config.ID(), Transport: parsed.Scheme}),
	}

	return &receiver, nil
}

func (r *Receiver) Start(ctx context.Context, host component.Host) error {
	go r.run()
	return nil
}

func (r *Receiver) Shutdown(ctx context.Context) error {
	return nil
}

func (r *Receiver) connect() (context.Context, error) {
	if r.podmanCtx != nil {
		fmt.Println("returning existing one")
		return r.podmanCtx, nil
	}

	ctx, err := bindings.NewConnection(context.Background(), r.config.Endpoint)
	if err == nil {
		r.podmanCtx = ctx
	}
	return ctx, err
}

/*
type result struct {
	md  *agentmetricspb.ExportMetricsServiceRequest
	err error
}
*/

func (r *Receiver) run() error {

	for {
		time.Sleep(3 * time.Second)
		conn, err := r.connect()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("got connection: ", conn)

		cntrs, err := containers.List(conn, &containers.ListOptions{})
		fmt.Println(cntrs, err)

		var mds []pdata.Metrics
		for _, md := range mds {
			err := r.nextConsumer.ConsumeMetrics(r.runnerCtx, md)
			fmt.Println(err)
		}

	}
	return nil

	/*
		if !r.successfullySetup {
			return r.Setup()
		}

		c := r.obsrecv.StartMetricsOp(r.obsCtx)

		containers := r.client.Containers()
		results := make(chan result, len(containers))

		wg := &sync.WaitGroup{}
		wg.Add(len(containers))
		for _, container := range containers {
			go func(dc DockerContainer) {
				md, err := r.client.FetchContainerStatsAndConvertToMetrics(r.runnerCtx, dc)
				results <- result{md, err}
				wg.Done()
			}(container)
		}

		wg.Wait()
		close(results)

		numPoints := 0
		var lastErr error
		for result := range results {
			var err error
			if result.md != nil {
				md := internaldata.OCToMetrics(result.md.Node, result.md.Resource, result.md.Metrics)
				_, np := md.MetricAndDataPointCount()
				numPoints += np
				err = r.nextConsumer.ConsumeMetrics(r.runnerCtx, md)
			} else {
				err = result.err
			}

			if err != nil {
				lastErr = err
			}
		}

		r.obsrecv.EndMetricsOp(c, typeStr, numPoints, lastErr)
		return nil
	*/
}
