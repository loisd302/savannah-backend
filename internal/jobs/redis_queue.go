package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/services"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	PendingJobsQueue = "sms_jobs:pending"
	ProcessingSet    = "sms_jobs:processing"
	CompletedSet     = "sms_jobs:completed"
	FailedSet        = "sms_jobs:failed"
	RetryQueue       = "sms_jobs:retry"
	JobDataPrefix    = "sms_job:"
	StatsPrefix      = "sms_stats:"
)

// RedisJobQueue implements JobQueue interface using Redis
type RedisJobQueue struct {
	client *redis.Client
}

// NewRedisJobQueue creates a new Redis-based job queue
func NewRedisJobQueue(client *redis.Client) *RedisJobQueue {
	return &RedisJobQueue{
		client: client,
	}
}

// Enqueue adds a job to the pending queue
func (r *RedisJobQueue) Enqueue(ctx context.Context, job *services.SMSJob) error {
	// Store job data
	jobKey := JobDataPrefix + job.ID.String()
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	pipe := r.client.TxPipeline()
	
	// Store job data with TTL (24 hours)
	pipe.Set(ctx, jobKey, jobData, 24*time.Hour)
	
	// Add to pending queue with priority score (timestamp)
	pipe.ZAdd(ctx, PendingJobsQueue, &redis.Z{
		Score:  float64(job.ScheduledFor.Unix()),
		Member: job.ID.String(),
	})
	
	// Update pending counter
	pipe.Incr(ctx, StatsPrefix+"pending")
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue retrieves and removes the next available job from the queue
func (r *RedisJobQueue) Dequeue(ctx context.Context) (*services.SMSJob, error) {
	// Get jobs that are ready to be processed (score <= current timestamp)
	now := time.Now().Unix()
	results, err := r.client.ZRangeByScore(ctx, PendingJobsQueue, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%d", now),
		Count: 1,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending job: %w", err)
	}

	if len(results) == 0 {
		// Check retry queue
		return r.dequeueFromRetry(ctx)
	}

	jobID := results[0]
	
	// Move job from pending to processing atomically
	pipe := r.client.TxPipeline()
	pipe.ZRem(ctx, PendingJobsQueue, jobID)
	pipe.SAdd(ctx, ProcessingSet, jobID)
	pipe.Decr(ctx, StatsPrefix+"pending")
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to move job to processing: %w", err)
	}

	// Fetch job data
	jobKey := JobDataPrefix + jobID
	jobData, err := r.client.Get(ctx, jobKey).Result()
	if err == redis.Nil {
		// Job data not found, clean up
		r.client.SRem(ctx, ProcessingSet, jobID)
		return nil, fmt.Errorf("job data not found for ID: %s", jobID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job data: %w", err)
	}

	var job services.SMSJob
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	return &job, nil
}

// dequeueFromRetry retrieves jobs from the retry queue
func (r *RedisJobQueue) dequeueFromRetry(ctx context.Context) (*services.SMSJob, error) {
	now := time.Now().Unix()
	results, err := r.client.ZRangeByScore(ctx, RetryQueue, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%d", now),
		Count: 1,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch retry job: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	jobID := results[0]
	
	// Move job from retry to processing
	pipe := r.client.TxPipeline()
	pipe.ZRem(ctx, RetryQueue, jobID)
	pipe.SAdd(ctx, ProcessingSet, jobID)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to move retry job to processing: %w", err)
	}

	// Fetch job data
	jobKey := JobDataPrefix + jobID
	jobData, err := r.client.Get(ctx, jobKey).Result()
	if err == redis.Nil {
		r.client.SRem(ctx, ProcessingSet, jobID)
		return nil, fmt.Errorf("retry job data not found for ID: %s", jobID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch retry job data: %w", err)
	}

	var job services.SMSJob
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal retry job data: %w", err)
	}

	return &job, nil
}

// UpdateJob updates job data in Redis
func (r *RedisJobQueue) UpdateJob(ctx context.Context, job *services.SMSJob) error {
	jobKey := JobDataPrefix + job.ID.String()
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job data: %w", err)
	}

	// Update job data with extended TTL
	err = r.client.Set(ctx, jobKey, jobData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update job data: %w", err)
	}

	return nil
}

// RetryJob schedules a job for retry
func (r *RedisJobQueue) RetryJob(ctx context.Context, job *services.SMSJob, delay time.Duration) error {
	job.ScheduledFor = time.Now().Add(delay)
	
	// Update job data
	if err := r.UpdateJob(ctx, job); err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	
	// Remove from processing
	pipe.SRem(ctx, ProcessingSet, job.ID.String())
	
	// Add to retry queue with delayed timestamp
	pipe.ZAdd(ctx, RetryQueue, &redis.Z{
		Score:  float64(job.ScheduledFor.Unix()),
		Member: job.ID.String(),
	})
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to schedule job for retry: %w", err)
	}

	return nil
}

// MarkCompleted marks a job as completed
func (r *RedisJobQueue) MarkCompleted(ctx context.Context, jobID uuid.UUID) error {
	pipe := r.client.TxPipeline()
	
	// Remove from processing and add to completed
	pipe.SRem(ctx, ProcessingSet, jobID.String())
	pipe.SAdd(ctx, CompletedSet, jobID.String())
	
	// Update stats
	pipe.Incr(ctx, StatsPrefix+"sent")
	
	// Set TTL for completed jobs (keep for 7 days)
	pipe.Expire(ctx, JobDataPrefix+jobID.String(), 7*24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark job as completed: %w", err)
	}

	return nil
}

// MarkFailed marks a job as failed
func (r *RedisJobQueue) MarkFailed(ctx context.Context, jobID uuid.UUID, errorMsg string) error {
	pipe := r.client.TxPipeline()
	
	// Remove from processing and add to failed
	pipe.SRem(ctx, ProcessingSet, jobID.String())
	pipe.SAdd(ctx, FailedSet, jobID.String())
	
	// Update stats
	pipe.Incr(ctx, StatsPrefix+"failed")
	
	// Store error message
	pipe.Set(ctx, "error:"+jobID.String(), errorMsg, 7*24*time.Hour)
	
	// Set TTL for failed jobs (keep for 7 days)
	pipe.Expire(ctx, JobDataPrefix+jobID.String(), 7*24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark job as failed: %w", err)
	}

	return nil
}

// GetStats returns job queue statistics
func (r *RedisJobQueue) GetStats(ctx context.Context) (map[string]int64, error) {
	pipe := r.client.Pipeline()
	
	pendingCmd := pipe.ZCard(ctx, PendingJobsQueue)
	retryCmd := pipe.ZCard(ctx, RetryQueue)
	processingCmd := pipe.SCard(ctx, ProcessingSet)
	completedCmd := pipe.SCard(ctx, CompletedSet)
	failedCmd := pipe.SCard(ctx, FailedSet)
	
	// Get counters
	sentCmd := pipe.Get(ctx, StatsPrefix+"sent")
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	sent, _ := sentCmd.Int64()
	if sentCmd.Err() == redis.Nil {
		sent = 0
	}

	return map[string]int64{
		"pending":    pendingCmd.Val(),
		"retry":      retryCmd.Val(),
		"processing": processingCmd.Val(),
		"completed":  completedCmd.Val(),
		"failed":     failedCmd.Val(),
		"sent":       sent,
	}, nil
}

// CleanupExpiredJobs removes old completed and failed jobs
func (r *RedisJobQueue) CleanupExpiredJobs(ctx context.Context) error {
	// This is handled by Redis TTL, but we can add additional cleanup logic here
	// For example, removing very old entries from sets
	
	cutoffTime := time.Now().AddDate(0, 0, -30).Unix() // 30 days ago
	
	pipe := r.client.TxPipeline()
	
	// Remove old entries from completed and failed sets
	// This would require storing timestamp info, so for now just return nil
	_ = cutoffTime
	_ = pipe
	
	return nil
}

// GetJobStatus returns the status of a specific job
func (r *RedisJobQueue) GetJobStatus(ctx context.Context, jobID uuid.UUID) (string, error) {
	jobIDStr := jobID.String()
	
	// Check each set to determine status
	isPending, err := r.client.ZScore(ctx, PendingJobsQueue, jobIDStr).Result()
	if err == nil {
		_ = isPending
		return "pending", nil
	}
	
	isRetry, err := r.client.ZScore(ctx, RetryQueue, jobIDStr).Result()
	if err == nil {
		_ = isRetry
		return "retry", nil
	}
	
	isProcessing, err := r.client.SIsMember(ctx, ProcessingSet, jobIDStr).Result()
	if err == nil && isProcessing {
		return "processing", nil
	}
	
	isCompleted, err := r.client.SIsMember(ctx, CompletedSet, jobIDStr).Result()
	if err == nil && isCompleted {
		return "completed", nil
	}
	
	isFailed, err := r.client.SIsMember(ctx, FailedSet, jobIDStr).Result()
	if err == nil && isFailed {
		return "failed", nil
	}
	
	return "unknown", nil
}