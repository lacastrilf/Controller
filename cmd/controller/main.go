package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"controller/internal/collector"
	appconfig "controller/internal/config"
	"controller/internal/controller"
	"controller/internal/evaluator"
	"controller/internal/executor"
	"controller/internal/logger"
	"controller/internal/utils"
)

type modeList []string

func (m *modeList) String() string {
	return fmt.Sprint(*m)
}

func (m *modeList) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func main() {
	var modes modeList
	flag.Var(&modes, "m", "mode to run (repeatable)")
	flag.Parse()

	modesConfig, err := utils.LoadYAML[appconfig.ModesConfig]("config/modes.yaml")
	if err != nil {
		fmt.Println("error loading modes.yaml:", err)
		return
	}

	infraConfig, err := utils.LoadYAML[appconfig.InfraConfig]("config/infra.yaml")
	if err != nil {
		fmt.Println("error loading infraestructura.yaml:", err)
		return
	}

	tiersInUse := make(map[string]string)
	for _, modeName := range modes {
		modeConfig, ok := modesConfig.Modes[modeName]
		if !ok {
			fmt.Printf("error: mode %q not found in modes.yaml\n", modeName)
			return
		}
		if existingMode, alreadyUsed := tiersInUse[modeConfig.Tier]; alreadyUsed {
			fmt.Printf("error: modes %q and %q both target tier %q — only one mode per tier can run at a time\n",
				existingMode, modeName, modeConfig.Tier)
			return
		}
		tiersInUse[modeConfig.Tier] = modeName
	}

	ctx := context.Background()

	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(infraConfig.VPC.Region))
	if err != nil {
		fmt.Println("error loading AWS credentials:", err)
		return
	}

	ec2Client := ec2.NewFromConfig(awsConfig)
	elbClient := elasticloadbalancingv2.NewFromConfig(awsConfig)
	cwClient := cloudwatch.NewFromConfig(awsConfig)

	logFile, err := os.Create("controller.jsonl")
	if err != nil {
		fmt.Println("error creating log file:", err)
		return
	}
	defer logFile.Close()

	log := logger.New(os.Stdout, logFile)

	var runners []*controller.Runner

	for _, modeName := range modes {
		modeConfig := modesConfig.Modes[modeName]
		tierInfra, ok := infraConfig.Tiers[modeConfig.Tier]
		if !ok {
			fmt.Printf("error: tier %q (used by mode %q) not found in infraestructura.yaml\n", modeConfig.Tier, modeName)
			return
		}

		var exec executor.Executor

		switch modeConfig.Tier {
		case "microserver":
			k8sExecutor, err := executor.NewKubernetesExecutor(
				infraConfig.Kubernetes.KubeconfigPath,
				infraConfig.Kubernetes.Namespace,
				infraConfig.Kubernetes.DeploymentName,
			)
			if err != nil {
				fmt.Printf("error creating Kubernetes executor for mode %s: %v\n", modeName, err)
				return
			}
			exec = k8sExecutor

		default: // application, server
			exec = &executor.EC2Executor{
				Client:          ec2Client,
				ELBClient:       elbClient,
				Tier:            modeConfig.Tier,
				AMIID:           tierInfra.AMIID,
				InstanceType:    tierInfra.InstanceType,
				KeyName:         tierInfra.KeyName,
				SubnetIDs:       tierInfra.Subnets,
				SecurityGroupID: tierInfra.SecurityGroupID,
				TargetGroupARN:  tierInfra.TargetGroupARN,
			}
		}

		metricsProvider, err := collectorForTier(modeConfig.Tier, tierInfra, cwClient)
		if err != nil {
			fmt.Printf("error creating metrics collector for mode %s: %v\n", modeName, err)
			return
		}

		rule := &evaluator.ModeRule{
			Config:          modeConfig,
			MetricsProvider: metricsProvider,
		}

		cooldown := time.Duration(modeConfig.CooldownSeconds) * time.Second
		if cooldown <= 0 {
			// Default: give a freshly-launched instance two evaluation
			// cycles' worth of time to show up in CloudWatch before the
			// runner is allowed to react to it again.
			cooldown = 2 * time.Duration(modeConfig.IntervalSeconds) * time.Second
		}

		runner := &controller.Runner{
			ModeName:    modeName,
			Rule:        rule,
			Executor:    exec,
			Logger:      log,
			Interval:    time.Duration(modeConfig.IntervalSeconds) * time.Second,
			Cooldown:    cooldown,
			MinCapacity: modeConfig.MinCapacity,
			MaxCapacity: modeConfig.MaxCapacity,
		}

		runners = append(runners, runner)
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for _, r := range runners {
		go r.Run(shutdownCtx)
	}

	fmt.Println("controller started, press Ctrl+C to stop")

	<-shutdownCtx.Done()
	fmt.Println("shutdown signal received, waiting for runners to finish...")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("controller stopped")
}

func collectorForTier(tier string, tierInfra appconfig.TierInfraConfig, cwClient *cloudwatch.Client) (collector.Provider, error) {
	switch tier {
	case "microserver":
		return nil, fmt.Errorf("metrics collection for tier %q is not implemented yet", tier)
	default:
		return &collector.CloudWatchCollector{
			Client:                cwClient,
			TargetGroupARNSuffix:  targetGroupARNSuffix(tierInfra.TargetGroupARN),
			LoadBalancerARNSuffix: targetGroupARNSuffix(tierInfra.LoadBalancerARN),
			LookbackWindow:        5 * time.Minute,
		}, nil
	}
}

func targetGroupARNSuffix(arn string) string {
	if idx := strings.LastIndex(arn, ":"); idx != -1 {
		return arn[idx+1:]
	}
	return arn
}
