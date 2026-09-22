package config

type InfraConfig struct {
	VPC          VPCConfig                  `yaml:"vpc"`
	PublicSubnet []string                   `yaml:"public_subnet"`
	AdminSubnet  string                     `yaml:"admin_subnet"`
	Tiers        map[string]TierInfraConfig `yaml:"tiers"`
	Kubernetes   KubernetesConfig           `yaml:"kubernetes"`
}

type VPCConfig struct {
	VPCID  string `yaml:"vpc_id"`
	Region string `yaml:"region"`
}
type TierInfraConfig struct {
	Subnets         []string `yaml:"subnets"`
	AMIID           string   `yaml:"ami_id,omitempty"`
	InstanceType    string   `yaml:"instance_type,omitempty"`
	KeyName         string   `yaml:"key_name,omitempty"`
	SecurityGroupID string   `yaml:"security_group_id,omitempty"`
	TargetGroupARN  string   `yaml:"target_group_arn,omitempty"`
	LoadBalancerARN string   `yaml:"load_balancer_arn"`
}
type KubernetesConfig struct {
	KubeconfigPath string `yaml:"kubeconfig_path"`
	Namespace      string `yaml:"namespace"`
	DeploymentName string `yaml:"deployment_name"`
	NodePort       int    `yaml:"node_port"`
}
