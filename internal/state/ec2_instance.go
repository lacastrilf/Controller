package state

import "time"

type InstanceStatus string

const (
	StatusPending     InstanceStatus = "pending"
	StatusRunning     InstanceStatus = "running"
	StatusTerminating InstanceStatus = "terminating"
)

type EC2Instance struct {
	ID         string
	Tier       string
	SubnetID   string
	PrivateIP  string
	Status     InstanceStatus
	LaunchedAt time.Time
}
