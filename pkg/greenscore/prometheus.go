package greenscore

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog"
)

type PrometheusHandle struct {
	timeRange time.Duration
	address   string
	api       v1.API
}

func NewPrometheus(address string, timeRange time.Duration) *PrometheusHandle {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		klog.Fatalf("[NetworkTraffic] Error creating prometheus client: %s", err.Error())
	}

	return &PrometheusHandle{
		timeRange: timeRange,
		address:   address,
		api:       v1.NewAPI(client),
	}
}

func (p *PrometheusHandle) GetPromInfo(node string) (*model.Sample, error) {
	res, err := p.query(query)
	if err != nil {
		return nil, fmt.Errorf("[NetworkTraffic] Error querying prometheus: %w", err)
	}

	nodeMeasure := res.(model.Vector)
	if len(nodeMeasure) != 1 {
		return nil, fmt.Errorf("[NetworkTraffic] Invalid response, expected 1 value, got %d", len(nodeMeasure))
	}

	return nodeMeasure[0], nil
}

func (p *PrometheusHandle) query(query string) (model.Value, error) {
	results, warnings, err := p.api.Query(context.Background(), query, time.Now())

	if len(warnings) > 0 {
		klog.Warningf("[NetworkTraffic] Warnings: %v\n", warnings)
	}

	return results, err
}
