package state

type PodStatus string

const (
	PodPending PodStatus = "Pending"
	PodRunning PodStatus = "Running"
	PodFailed  PodStatus = "Failed"
)

type PodReplica struct {
	ID     string
	Status PodStatus
}

type MicroserverState struct {
	DesiredReplicas int
	ReadyReplicas   int
	Replicas        []PodReplica
}
