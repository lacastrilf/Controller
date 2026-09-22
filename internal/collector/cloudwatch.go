package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type CloudWatchCollector struct {
	Client                    *cloudwatch.Client
	TargetGroupARNSuffix      string
	LoadBalancerARNSuffix     string
	QueueLengthNamespace      string
	QueueLengthDimensionName  string
	QueueLengthDimensionValue string
	LookbackWindow            time.Duration
}

func (c *CloudWatchCollector) GetMetric(ctx context.Context, resourceIDs []string, metricName string) ([]Sample, error) {
	switch metricName {
	case "cpu":
		return c.getPerInstanceMetric(ctx, resourceIDs, "AWS/EC2", "CPUUtilization", "InstanceId")
	case "memory":
		return c.getPerInstanceMetric(ctx, resourceIDs, "CWAgent", "mem_used_percent", "InstanceId")
	case "response_time":
		return c.getALBMetric(ctx)
	case "queue_length":
		return c.getCustomMetric(ctx)
	default:
		return nil, fmt.Errorf("unknown metric: %s", metricName)
	}
}

func (c *CloudWatchCollector) getPerInstanceMetric(ctx context.Context, instanceIDs []string, namespace, metricName, dimensionName string) ([]Sample, error) {
	if len(instanceIDs) == 0 {
		return nil, fmt.Errorf("no instance IDs provided for metric %s", metricName)
	}

	end := time.Now()
	start := end.Add(-c.LookbackWindow)

	var allSamples [][]Sample
	for _, id := range instanceIDs {
		samples, err := c.queryMetric(ctx, namespace, metricName, dimensionName, id, start, end)
		if err != nil {
			return nil, err
		}
		allSamples = append(allSamples, samples)
	}

	return averageSamplesAcrossInstances(allSamples), nil
}

func (c *CloudWatchCollector) getALBMetric(ctx context.Context) ([]Sample, error) {
	end := time.Now()
	start := end.Add(-c.LookbackWindow)
	return c.queryMetric(ctx, "AWS/ApplicationELB", "TargetResponseTime", "TargetGroup", c.TargetGroupARNSuffix, start, end)
}

func (c *CloudWatchCollector) getCustomMetric(ctx context.Context) ([]Sample, error) {
	end := time.Now()
	start := end.Add(-c.LookbackWindow)
	return c.queryMetric(ctx, c.QueueLengthNamespace, "queue_length", c.QueueLengthDimensionName, c.QueueLengthDimensionValue, start, end)
}

func (c *CloudWatchCollector) queryMetric(ctx context.Context, namespace, metricName, dimensionName, dimensionValue string, start, end time.Time) ([]Sample, error) {
	output, err := c.Client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: []cwtypes.Dimension{
			{Name: aws.String(dimensionName), Value: aws.String(dimensionValue)},
		},
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Period:     aws.Int32(60), // 1-minute resolution
		Statistics: []cwtypes.Statistic{cwtypes.StatisticAverage},
	})
	if err != nil {
		return nil, fmt.Errorf("querying %s/%s: %w", namespace, metricName, err)
	}

	samples := make([]Sample, len(output.Datapoints))
	for i, dp := range output.Datapoints {
		samples[i] = Sample{
			Timestamp: aws.ToTime(dp.Timestamp),
			Value:     aws.ToFloat64(dp.Average),
		}
	}
	return samples, nil
}

func averageSamplesAcrossInstances(allSamples [][]Sample) []Sample {
	if len(allSamples) == 0 {
		return nil
	}

	shortest := allSamples[0]
	for _, s := range allSamples {
		if len(s) < len(shortest) {
			shortest = s
		}
	}

	combined := make([]Sample, len(shortest))
	for i := range shortest {
		sum := 0.0
		for _, series := range allSamples {
			sum += series[i].Value
		}
		combined[i] = Sample{
			Timestamp: shortest[i].Timestamp,
			Value:     sum / float64(len(allSamples)),
		}
	}
	return combined
}
