package docker

const (
	StatusOK = iota

	StatusPaused  = "paused"
	StatusRemove  = "removing"
	StatusCreated = "created"
	StatusRunning = "running"
	StatusRestart = "restarting"
	StatusExited  = "exited"
	StatusDead    = "dead"
)
