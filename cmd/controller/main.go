package main

import (
	"controller/internal/config"
	"controller/internal/utils"
	"fmt"
)

func main() {
	modes, err := utils.LoadYAML[config.ModesConfig]("config/modes.yaml")
	if err != nil {
		fmt.Println("error cargando modes.yaml:", err)
		return
	}

	infra, err := utils.LoadYAML[config.InfraConfig]("config/infra.yaml")
	if err != nil {
		fmt.Println("error cargando infraestructura.yaml:", err)
		return
	}

	fmt.Println("--- Modos ---")
	fmt.Println("modos cargados:", len(modes.Modes))
	for nombre, modo := range modes.Modes {
		fmt.Println("modo:", nombre, "| métricas:", len(modo.Metrics))
	}

	fmt.Println("\n--- Infraestructura ---")
	fmt.Println("VPC:", infra.VPC.VPCID, "| región:", infra.VPC.Region)
	fmt.Println("subred application:", infra.Subnets.Application)
	fmt.Println("subred server:", infra.Subnets.Server)
	fmt.Println("subred microserver:", infra.Subnets.Microserver)
	fmt.Println("AMI application:", infra.EC2.Application.AMIID)
	fmt.Println("AMI server:", infra.EC2.Server.AMIID)
	fmt.Println("target group application:", infra.LoadBalancer.ApplicationTargetGroupARN)
	fmt.Println("kubeconfig:", infra.Kubernetes.KubeconfigPath, "| node_port:", infra.Kubernetes.NodePort)
}
