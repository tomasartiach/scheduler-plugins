package greenscore

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	pluginconfig "sigs.k8s.io/scheduler-plugins/apis/config"
)

type GreenScore struct {
	handle     framework.Handle
	prometheus *PrometheusHandle
}

const Name = "GreenScore"

var _ = framework.ScorePlugin(&GreenScore{})

// New initializes a new plugin and returns it.
func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args, ok := obj.(*pluginconfig.GreenScoreArgs)
	if !ok {
		return nil, fmt.Errorf("[GreenScore] want args to be of type NetworkTrafficArgs, got %T", obj)
	}

	klog.Infof("[GreenScore] args received. TimeRangeInMinutes: %d, Address: %s", args.TimeRangeInMinutes, args.Address)

	return &GreenScore{
		handle:     h,
		prometheus: NewPrometheus(args.Address, time.Minute*time.Duration(args.TimeRangeInMinutes)),
	}, nil
}

// Name returns name of the plugin. It is used in logs, etc.
func (n *GreenScore) Name() string {
	return Name
}

func (n *GreenScore) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeBandwidth, err := n.prometheus.GetPromInfo(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("error getting node bandwidth measure: %s", err))
	}

	klog.Infof("[GreenScore] node '%s' bandwidth: %s", nodeName, nodeBandwidth.Value)
	return int64(nodeBandwidth.Value), nil
}

func (n *GreenScore) ScoreExtensions() framework.ScoreExtensions {
	return n
}

func (n *GreenScore) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var higherScore int64
	for _, node := range scores {
		if higherScore < node.Score {
			higherScore = node.Score
		}
	}

	for i, node := range scores {
		scores[i].Score = framework.MaxNodeScore - (node.Score * framework.MaxNodeScore / higherScore)
	}

	klog.Infof("[GreenScore] Nodes final score: %v", scores)
	return nil
}
