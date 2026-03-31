package stream_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/llm"
	"Alice088/essentia/internal/domain/pipeline"
	"Alice088/essentia/internal/llm_manager"
	"Alice088/essentia/pkg/s3"
	"Alice088/essentia/pkg/storage"
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockStorage implements storage.Storage for testing.
type mockStorage struct {
	mu               sync.RWMutex
	jobsToReturn     []storage.Job
	invalidJobCalled []uuid.UUID
}

func (m *mockStorage) GetProcessableJobs(ctx context.Context, limit int) []storage.Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobsToReturn
}

func (m *mockStorage) InvalidJob(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invalidJobCalled = append(m.invalidJobCalled, id)
	return nil
}

func (m *mockStorage) setJobs(jobs []storage.Job) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobsToReturn = jobs
}

func (m *mockStorage) getInvalidJobCalls() []uuid.UUID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.invalidJobCalled
}

// mockS3 implements s3.S3 for testing.
type mockS3 struct {
	mu           sync.RWMutex
	findResults  map[string][]pipeline.Blob // key: id+bucket
	findErrors   map[string]error
	putCalled    []s3.File
	getCalled    []s3.File
	deleteCalled []s3.File
}

func (m *mockS3) Put(ctx context.Context, file s3.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.putCalled = append(m.putCalled, file)
	return nil
}

func (m *mockS3) Get(ctx context.Context, file s3.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalled = append(m.getCalled, file)
	return nil
}

func (m *mockS3) Find(ctx context.Context, name, bucket string) ([]pipeline.Blob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name + ":" + bucket
	if err, ok := m.findErrors[key]; ok {
		return nil, err
	}
	if result, ok := m.findResults[key]; ok {
		return result, nil
	}
	return []pipeline.Blob{"mock-blob"}, nil
}

func (m *mockS3) Delete(ctx context.Context, file s3.File) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalled = append(m.deleteCalled, file)
	return nil
}

func (m *mockS3) setFindResult(name, bucket string, blobs []pipeline.Blob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.findResults == nil {
		m.findResults = make(map[string][]pipeline.Blob)
	}
	m.findResults[name+":"+bucket] = blobs
}

func (m *mockS3) setFindError(name, bucket string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.findErrors == nil {
		m.findErrors = make(map[string]error)
	}
	m.findErrors[name+":"+bucket] = err
}

// testStreamManager creates a StreamManager with test doubles.
func testStreamManager(t *testing.T, cfg config.StreamManager, llmManager *llm_manager.Manager) (*StreamManager, *mockStorage, *mockS3, map[string]chan pipeline.Job) {
	t.Helper()
	storageMock := &mockStorage{}
	s3Mock := &mockS3{}
	streams := make(map[string]chan pipeline.Job)
	// Create a stream for testing
	streams["test-stage"] = make(chan pipeline.Job, 10)

	logger := slog.Default()

	return &StreamManager{
		Config:  cfg,
		Streams: streams,
		Storage: storageMock,
		S3:      s3Mock,
		LLM:     llmManager,
		Logger:  logger,
	}, storageMock, s3Mock, streams
}

func TestStreamManager_LLMSoftStop_BlocksNewJobs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create LLM manager in soft stop state
	llmCfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &llm_manager.StubBalanceProvider{Balance: 0.30} // Below soft limit
	llmManager := llm_manager.NewWithBalance(llmCfg, provider)
	if err := llmManager.UpdateBalance(); err != nil {
		t.Fatal(err)
	}
	snap := llmManager.Snapshot()
	if snap.State != llm.LimitStateSoftStop {
		t.Fatalf("Expected LLM state SoftStop, got %s", snap.State)
	}

	// StreamManager config
	streamCfg := config.StreamManager{
		Ticker:         50 * time.Millisecond,
		JobBatchCount:  5,
		JobPullTimeout: 1 * time.Second,
	}

	sm, storageMock, s3Mock, streams := testStreamManager(t, streamCfg, llmManager)

	// Create a test job
	jobID := uuid.New()
	storageJob := storage.Job{
		ID:    jobID,
		Stage: "test-stage",
	}
	storageMock.setJobs([]storage.Job{storageJob})
	s3Mock.setFindResult(jobID.String(), "test-stage", []pipeline.Blob{"data1", "data2"})

	// Start manager
	go sm.Manage(ctx)

	// Wait for a few ticks
	time.Sleep(200 * time.Millisecond)

	// Check that no job was sent to stream (SoftStop should block)
	select {
	case job := <-streams["test-stage"]:
		t.Errorf("Expected no job dispatched in SoftStop state, got job %s", job.JobID)
	default:
		// Good, no job dispatched
	}

	// Verify that Find was called (job was prepared)
	// but PullJobs should have skipped dispatch due to SoftStop
	// We can't directly observe that, but we can verify that the job wasn't invalidated
	invalidCalls := storageMock.getInvalidJobCalls()
	if len(invalidCalls) > 0 {
		t.Errorf("Expected no invalid job calls, got %d", len(invalidCalls))
	}
}

func TestStreamManager_LLMMaxStop_EmergencyShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create LLM manager in max stop state
	llmCfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &llm_manager.StubBalanceProvider{Balance: 0.05} // Below max limit
	llmManager := llm_manager.NewWithBalance(llmCfg, provider)
	if err := llmManager.UpdateBalance(); err != nil {
		t.Fatal(err)
	}
	snap := llmManager.Snapshot()
	if snap.State != llm.LimitStateMaxStop {
		t.Fatalf("Expected LLM state MaxStop, got %s", snap.State)
	}

	// StreamManager config with fast ticker
	streamCfg := config.StreamManager{
		Ticker:         1 * time.Millisecond,
		JobBatchCount:  5,
		JobPullTimeout: 1 * time.Second,
	}

	sm, _, _, _ := testStreamManager(t, streamCfg, llmManager)

	// Use a WaitGroup to detect when Manage goroutine exits
	var wg sync.WaitGroup
	wg.Add(1)
	sm.wg = &wg

	// Start manager
	go sm.Manage(ctx)

	// Wait for Manage goroutine to exit (due to emergency shutdown)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Goroutine exited, emergency shutdown successful
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for emergency shutdown")
	}

	// Ensure parent context is still active (shutdown wasn't due to parent cancellation)
	if ctx.Err() != nil {
		t.Error("Parent context should not be cancelled yet")
	}
}

func TestStreamManager_LLMNormal_JobsDispatched(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// LLM manager in normal state
	llmCfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &llm_manager.StubBalanceProvider{Balance: 1.00} // Above soft limit
	llmManager := llm_manager.NewWithBalance(llmCfg, provider)
	if err := llmManager.UpdateBalance(); err != nil {
		t.Fatal(err)
	}
	snap := llmManager.Snapshot()
	if snap.State != llm.LimitStateNormal {
		t.Fatalf("Expected LLM state Normal, got %s", snap.State)
	}

	streamCfg := config.StreamManager{
		Ticker:         50 * time.Millisecond,
		JobBatchCount:  5,
		JobPullTimeout: 1 * time.Second,
	}

	sm, storageMock, s3Mock, streams := testStreamManager(t, streamCfg, llmManager)

	// Create test jobs
	jobID1 := uuid.New()
	jobID2 := uuid.New()
	storageJobs := []storage.Job{
		{ID: jobID1, Stage: "test-stage"},
		{ID: jobID2, Stage: "test-stage"},
	}
	storageMock.setJobs(storageJobs)
	s3Mock.setFindResult(jobID1.String(), "test-stage", []pipeline.Blob{"data1"})
	s3Mock.setFindResult(jobID2.String(), "test-stage", []pipeline.Blob{"data2"})

	// Start manager
	go sm.Manage(ctx)

	// Collect dispatched jobs
	var receivedJobs []uuid.UUID
	for i := 0; i < 2; i++ {
		select {
		case job := <-streams["test-stage"]:
			receivedJobs = append(receivedJobs, job.JobID)
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("Timeout waiting for job %d", i+1)
		}
	}

	// Verify both jobs were dispatched
	if len(receivedJobs) != 2 {
		t.Errorf("Expected 2 jobs dispatched, got %d", len(receivedJobs))
	}
	// Check that both job IDs are present
	idSet := make(map[uuid.UUID]bool)
	for _, id := range receivedJobs {
		idSet[id] = true
	}
	if !idSet[jobID1] || !idSet[jobID2] {
		t.Errorf("Not all expected jobs were dispatched: got %v", receivedJobs)
	}
}

func TestStreamManager_PullJobs_RespectsLLMState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamCfg := config.StreamManager{
		Ticker:         time.Minute, // Not used in this test
		JobBatchCount:  5,
		JobPullTimeout: 1 * time.Second,
	}

	// Test cases for different LLM states
	tests := []struct {
		name       string
		balance    float64
		expectSend bool
		expectStop bool
	}{
		{
			name:       "normal_state_sends",
			balance:    1.00,
			expectSend: true,
			expectStop: false,
		},
		{
			name:       "soft_stop_blocks",
			balance:    0.30,
			expectSend: false,
			expectStop: false,
		},
		{
			name:       "max_stop_shutsdown",
			balance:    0.05,
			expectSend: false,
			expectStop: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmCfg := config.LLMManager{
				Enabled:          true,
				SoftBalanceLimit: 0.50,
				MaxBalanceLimit:  0.10,
			}
			provider := &llm_manager.StubBalanceProvider{Balance: tt.balance}
			llmManager := llm_manager.NewWithBalance(llmCfg, provider)
			if err := llmManager.UpdateBalance(); err != nil {
				t.Fatal(err)
			}

			sm, _, _, streams := testStreamManager(t, streamCfg, llmManager)

			// Create a test job
			job := pipeline.Job{
				JobID:      uuid.New(),
				Stage:      "test-stage",
				Input:      []pipeline.Blob{"test"},
				LLMContext: llm.Context{Prompt: 100, Completion: 50},
			}

			// Monitor for shutdown
			var shutdownCalled atomic.Bool
			originalStop := sm.stop
			sm.stop = func() {
				shutdownCalled.Store(true)
				if originalStop != nil {
					originalStop()
				}
			}

			// Call PullJobs directly
			go sm.PullJobs(ctx, job)

			// Check result
			timeout := time.After(100 * time.Millisecond)
			if tt.expectSend {
				select {
				case receivedJob := <-streams["test-stage"]:
					if receivedJob.JobID != job.JobID {
						t.Errorf("Received wrong job ID: %s", receivedJob.JobID)
					}
				case <-timeout:
					t.Error("Expected job to be sent to stream")
				}
			} else {
				select {
				case <-streams["test-stage"]:
					t.Error("Job should not be sent to stream")
				case <-timeout:
					// Good, no job sent
				}
			}

			if tt.expectStop && !shutdownCalled.Load() {
				t.Error("Expected shutdown to be called")
			}
			if !tt.expectStop && shutdownCalled.Load() {
				t.Error("Unexpected shutdown")
			}
		})
	}
}
