package executor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type EC2Executor struct {
	Client          *ec2.Client
	ELBClient       *elasticloadbalancingv2.Client
	Tier            string // e.g. "application" — used as the Tier tag value
	AMIID           string
	InstanceType    string
	KeyName         string
	SubnetIDs       []string // one per AZ — a new instance picks one round-robin
	SecurityGroupID string
	TargetGroupARN  string
}

func (e *EC2Executor) describeTierInstances(ctx context.Context) ([]ec2types.Instance, error) {
	output, err := e.Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:Tier"),
				Values: []string{e.Tier},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("describing instances for tier %s: %w", e.Tier, err)
	}

	var instances []ec2types.Instance
	for _, reservation := range output.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	return instances, nil
}

func (e *EC2Executor) CurrentCapacity(ctx context.Context) (int, error) {
	instances, err := e.describeTierInstances(ctx)
	if err != nil {
		return 0, err
	}
	return len(instances), nil
}

func (e *EC2Executor) CurrentResourceIDs(ctx context.Context) ([]string, error) {
	instances, err := e.describeTierInstances(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(instances))
	for i, instance := range instances {
		ids[i] = aws.ToString(instance.InstanceId)
	}
	return ids, nil
}

func (e *EC2Executor) ScaleUp(ctx context.Context, count int) error {
	subnetID := e.SubnetIDs[0]

	runOutput, err := e.Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:          aws.String(e.AMIID),
		InstanceType:     ec2types.InstanceType(e.InstanceType),
		KeyName:          aws.String(e.KeyName),
		SubnetId:         aws.String(subnetID),
		SecurityGroupIds: []string{e.SecurityGroupID},
		MinCount:         aws.Int32(int32(count)),
		MaxCount:         aws.Int32(int32(count)),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags: []ec2types.Tag{
					{Key: aws.String("Tier"), Value: aws.String(e.Tier)},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("launching %d instance(s) for tier %s: %w", count, e.Tier, err)
	}

	var instanceIDs []string
	for _, instance := range runOutput.Instances {
		instanceIDs = append(instanceIDs, aws.ToString(instance.InstanceId))
	}

	waiter := ec2.NewInstanceRunningWaiter(e.Client)
	err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	}, 90*time.Second)
	if err != nil {
		return fmt.Errorf("launched %d instance(s) for tier %s but they never reached running state: %w", count, e.Tier, err)
	}

	var targets []elbv2types.TargetDescription
	for _, id := range instanceIDs {
		targets = append(targets, elbv2types.TargetDescription{Id: aws.String(id)})
	}

	_, err = e.ELBClient.RegisterTargets(ctx, &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(e.TargetGroupARN),
		Targets:        targets,
	})
	if err != nil {
		return fmt.Errorf("launched %d instance(s) for tier %s but failed to register them: %w", count, e.Tier, err)
	}
	return e.waitUntilHealthy(ctx, targets, 2*time.Minute)
}

func (e *EC2Executor) registerTargets(ctx context.Context, instanceIDs []string) error {
	if len(instanceIDs) == 0 {
		return nil
	}

	targets := make([]elbv2types.TargetDescription, len(instanceIDs))
	for i, id := range instanceIDs {
		targets[i] = elbv2types.TargetDescription{Id: aws.String(id)}
	}

	_, err := e.ELBClient.RegisterTargets(ctx, &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(e.TargetGroupARN),
		Targets:        targets,
	})
	if err != nil {
		return fmt.Errorf("registering targets in %s: %w", e.TargetGroupARN, err)
	}
	return nil
}

func (e *EC2Executor) deregisterTargets(ctx context.Context, instanceIDs []string) error {
	if len(instanceIDs) == 0 {
		return nil
	}

	targets := make([]elbv2types.TargetDescription, len(instanceIDs))
	for i, id := range instanceIDs {
		targets[i] = elbv2types.TargetDescription{Id: aws.String(id)}
	}

	_, err := e.ELBClient.DeregisterTargets(ctx, &elasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(e.TargetGroupARN),
		Targets:        targets,
	})
	if err != nil {
		return fmt.Errorf("deregistering targets from %s: %w", e.TargetGroupARN, err)
	}
	return nil
}

func (e *EC2Executor) ScaleDown(ctx context.Context, count int) error {
	instances, err := e.describeTierInstances(ctx)
	if err != nil {
		return err
	}
	if len(instances) < count {
		return fmt.Errorf("cannot scale down by %d: only %d instance(s) running", count, len(instances))
	}

	// Sort newest first, so the newest instances are terminated first.
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].LaunchTime.After(*instances[j].LaunchTime)
	})

	idsToTerminate := make([]string, count)
	for i := 0; i < count; i++ {
		idsToTerminate[i] = aws.ToString(instances[i].InstanceId)
	}

	if err := e.deregisterTargets(ctx, idsToTerminate); err != nil {
		return fmt.Errorf("cannot scale down %d instance(s) for tier %s: %w", count, e.Tier, err)
	}

	_, err = e.Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: idsToTerminate,
	})
	if err != nil {
		return fmt.Errorf("terminating %d instance(s) for tier %s: %w", count, e.Tier, err)
	}
	return nil
}

func (e *EC2Executor) waitUntilHealthy(ctx context.Context, targets []elbv2types.TargetDescription, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		output, err := e.ELBClient.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(e.TargetGroupARN),
			Targets:        targets,
		})
		if err != nil {
			return fmt.Errorf("checking target health: %w", err)
		}

		allHealthy := true
		for _, desc := range output.TargetHealthDescriptions {
			if desc.TargetHealth.State != elbv2types.TargetHealthStateEnumHealthy {
				allHealthy = false
				break
			}
		}
		if allHealthy {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// keep polling
		}
	}

	return fmt.Errorf("target(s) did not become healthy within %s", timeout)
}
