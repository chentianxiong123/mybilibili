package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Scheduler runs enabled scheduled tasks.
type Scheduler struct {
	svc  *Service
	quit chan struct{}
}

func NewScheduler(svc *Service) *Scheduler {
	return &Scheduler{svc: svc, quit: make(chan struct{})}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.checkTasks(ctx)
		case <-s.quit:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.quit)
}

func (s *Scheduler) checkTasks(ctx context.Context) {
	tasks, err := s.svc.ListScheduledTasks(ctx)
	if err != nil {
		return
	}
	now := time.Now()
	for _, t := range tasks {
		if t.Enabled == 0 {
			continue
		}
		if t.NextRunAt != nil && t.NextRunAt.After(now) {
			continue
		}
		go s.runTask(ctx, t)
	}
}

func (s *Scheduler) runTask(ctx context.Context, t *ScheduledTask) {
	result := "success"
	message := ""

	// Default next run: 24h later
	nextRun := time.Now().Add(24 * time.Hour)

	switch t.TaskType {
	case "hot_search_cleanup":
		// Call search-service hot cleanup endpoint
		client := &http.Client{Timeout: 10 * time.Second}
		searchAddr := os.Getenv("SEARCH_SERVICE_ADDR")
		if searchAddr == "" {
			searchAddr = "localhost:8084"
		}
		url := "http://" + searchAddr + "/api/v1/search/hot/clean-expired"
		resp, err := client.Post(url, "application/json", nil)
		if err != nil {
			result = "failed"
			message = "cleanup request failed: " + err.Error()
		} else {
			resp.Body.Close()
			if resp.StatusCode != 200 {
				result = "failed"
				message = fmt.Sprintf("cleanup returned status %d", resp.StatusCode)
			} else {
				message = "hot search expired entries cleaned"
			}
		}
		// 每天凌晨3点
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		nextRun = next
	default:
		result = "skipped"
		message = "unknown task type: " + t.TaskType
	}

	if err := s.svc.UpdateScheduledTaskRun(ctx, t.TaskKey, result, message, &nextRun); err != nil {
		log.Printf("scheduler: failed to update task %s: %v", t.TaskKey, err)
	}
}

// TriggerTaskNow immediately runs a task by key.
func (s *Scheduler) TriggerTaskNow(ctx context.Context, taskKey string) error {
	t, err := s.svc.GetScheduledTaskByKey(ctx, taskKey)
	if err != nil {
		return err
	}
	s.runTask(ctx, t)
	return nil
}