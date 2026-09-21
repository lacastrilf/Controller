package config

type InfraConfig struct {
	VPC          VPCConfig          `yaml:"vpc"`
	Subnets      SubnetsConfig      `yaml:"subnets"`
	EC2          EC2Config          `yaml:"ec2"`
	LoadBalancer LoadBalancerConfig `yaml:"load_balancer"`
	Kubernetes   KubernetesConfig   `yaml:"kubernetes"`
}

type VPCConfig struct {
	VPCID  string `yaml:"vpc_id"`
	Region string `yaml:"region"`
}

type SubnetsConfig struct {
	Public            string `yaml:"public"`
	PrivateController string `yaml:"private_controller"`
	Application       string `yaml:"application"`
	Server            string `yaml:"server"`
	Microserver       string `yaml:"microserver"`
}

type EC2Config struct {
	Application EC2InstanceConfig `yaml:"application"`
	Server      EC2InstanceConfig `yaml:"server"`
}

type EC2InstanceConfig struct {
	AMIID           string `yaml:"ami_id"`
	InstanceType    string `yaml:"instance_type"`
	KeyName         string `yaml:"key_name"`
	SecurityGroupID string `yaml:"security_group_id"`
}

type LoadBalancerConfig struct {
	ApplicationTargetGroupARN string `yaml:"application_target_group_arn"`
	ServerTargetGroupARN      string `yaml:"server_target_group_arn"`
}

type KubernetesConfig struct {
	KubeconfigPath string `yaml:"kubeconfig_path"`
	Namespace      string `yaml:"namespace"`
	DeploymentName string `yaml:"deployment_name"`
	NodePort       int    `yaml:"node_port"`
}
