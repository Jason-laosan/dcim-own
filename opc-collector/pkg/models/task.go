package models

import "time"

// CollectionTask represents a task to collect data from a device
type CollectionTask struct {
	TaskID     string        `json:"task_id"`
	DeviceID   string        `json:"device_id"`
	Device     *OPCDevice    `json:"device"`
	Interval   time.Duration `json:"interval"`
	NextRun    time.Time     `json:"next_run"`
	Priority   int           `json:"priority"`
	Enabled    bool          `json:"enabled"`
	LastRun    time.Time     `json:"last_run"`
	LastStatus string        `json:"last_status"`
	RunCount   int64         `json:"run_count"`
	ErrorCount int64         `json:"error_count"`
}

// ShouldRun returns true if the task should be executed now
func (t *CollectionTask) ShouldRun() bool {
	return t.Enabled && time.Now().After(t.NextRun)
}

// UpdateNextRun calculates and updates the next run time
func (t *CollectionTask) UpdateNextRun() {
	t.NextRun = time.Now().Add(t.Interval)
}

// RecordSuccess records a successful execution
func (t *CollectionTask) RecordSuccess() {
	t.LastRun = time.Now()
	t.LastStatus = "success"
	t.RunCount++
	t.UpdateNextRun()
}

// RecordFailure records a failed execution
func (t *CollectionTask) RecordFailure() {
	t.LastRun = time.Now()
	t.LastStatus = "failure"
	t.ErrorCount++
	t.UpdateNextRun()
}
