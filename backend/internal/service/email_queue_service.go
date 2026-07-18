package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// Task type constants
const (
	TaskTypeVerifyCode    = "verify_code"
	TaskTypePasswordReset = "password_reset"
)

// EmailTask 邮件发送任务
type EmailTask struct {
	Email    string
	SiteName string
	TaskType string
	ResetURL string // Only used for password_reset task type
	Locale   string // Optional Accept-Language locale hint
}

// EmailQueueService 异步邮件队列服务
type EmailQueueService struct {
	emailService          *EmailService
	announcementBatchRepo AnnouncementEmailBatchRepository
	taskChan              chan EmailTask
	announcementWake      chan struct{}
	wg                    sync.WaitGroup
	stopChan              chan struct{}
	workers               int
}

// NewEmailQueueService 创建邮件队列服务
func NewEmailQueueService(emailService *EmailService, workers int) *EmailQueueService {
	return newEmailQueueService(emailService, nil, workers)
}

func NewPersistentEmailQueueService(emailService *EmailService, batchRepo AnnouncementEmailBatchRepository, workers int) *EmailQueueService {
	return newEmailQueueService(emailService, batchRepo, workers)
}

func newEmailQueueService(emailService *EmailService, batchRepo AnnouncementEmailBatchRepository, workers int) *EmailQueueService {
	if workers <= 0 {
		workers = 3 // 默认3个工作协程
	}

	service := &EmailQueueService{
		emailService:          emailService,
		announcementBatchRepo: batchRepo,
		taskChan:              make(chan EmailTask, 100), // 缓冲100个任务
		announcementWake:      make(chan struct{}, 1),
		stopChan:              make(chan struct{}),
		workers:               workers,
	}

	// 启动工作协程
	service.start()
	service.WakeAnnouncementWorker()

	return service
}

// start 启动工作协程
func (s *EmailQueueService) start() {
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	if s.announcementBatchRepo != nil {
		s.wg.Add(1)
		go s.announcementWorker()
	}
	logger.LegacyPrintf("service.email_queue", "[EmailQueue] Started %d workers", s.workers)
}

func (s *EmailQueueService) announcementWorker() {
	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.announcementWake:
			s.processDueAnnouncementBatches()
		case <-ticker.C:
			s.processDueAnnouncementBatches()
		case <-s.stopChan:
			return
		}
	}
}

func (s *EmailQueueService) processDueAnnouncementBatches() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		batch, err := s.announcementBatchRepo.ClaimDue(ctx, 10*time.Minute)
		cancel()
		if err != nil {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Failed to claim announcement email batch: %v", err)
			return
		}
		if batch == nil {
			return
		}
		s.processPersistentAnnouncementBatch(batch)
	}
}

func (s *EmailQueueService) processPersistentAnnouncementBatch(batch *AnnouncementEmailBatch) {
	heartbeatDone := make(chan struct{})
	heartbeatStopped := make(chan struct{})
	go s.refreshAnnouncementBatchLock(batch.ID, heartbeatDone, heartbeatStopped)

	processed, failed, sendErr := s.sendAnnouncementRecipients(batch.AnnouncementID, batch.CampaignID, batch.Recipients, batch.Title, batch.Content)
	close(heartbeatDone)
	<-heartbeatStopped

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if sendErr == nil {
		if err := s.announcementBatchRepo.MarkCompleted(ctx, batch.ID, processed); err != nil {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Failed to complete announcement email batch %d: %v", batch.ID, err)
		}
		return
	}

	nextAttemptAt := time.Now().Add(announcementEmailRetryDelay(batch.AttemptCount))
	lastError := sendErr.Error()
	if len(lastError) > 2000 {
		lastError = lastError[:2000]
	}
	if err := s.announcementBatchRepo.MarkRetry(ctx, batch.ID, processed, failed, lastError, nextAttemptAt); err != nil {
		logger.LegacyPrintf("service.email_queue", "[EmailQueue] Failed to reschedule announcement email batch %d: %v", batch.ID, err)
	}
}

func (s *EmailQueueService) refreshAnnouncementBatchLock(batchID int64, done <-chan struct{}, stopped chan<- struct{}) {
	defer close(stopped)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := s.announcementBatchRepo.RefreshLock(ctx, batchID)
			cancel()
			if err != nil {
				logger.LegacyPrintf("service.email_queue", "[EmailQueue] Failed to refresh announcement email batch %d lock: %v", batchID, err)
			}
		case <-done:
			return
		}
	}
}

func announcementEmailRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Minute << min(attempt-1, 6)
	if delay > time.Hour {
		return time.Hour
	}
	return delay
}

// worker 工作协程
func (s *EmailQueueService) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case task := <-s.taskChan:
			s.processTask(id, task)
		case <-s.stopChan:
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d stopping", id)
			return
		}
	}
}

// processTask 处理任务
func (s *EmailQueueService) processTask(workerID int, task EmailTask) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch task.TaskType {
	case TaskTypeVerifyCode:
		if err := s.emailService.SendVerifyCode(ctx, task.Email, task.SiteName, task.Locale); err != nil {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d failed to send verify code to %s: %v", workerID, task.Email, err)
		} else {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d sent verify code to %s", workerID, task.Email)
		}
	case TaskTypePasswordReset:
		if err := s.emailService.SendPasswordResetEmailWithCooldown(ctx, task.Email, task.SiteName, task.ResetURL, task.Locale); err != nil {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d failed to send password reset to %s: %v", workerID, task.Email, err)
		} else {
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d sent password reset to %s", workerID, task.Email)
		}
	default:
		logger.LegacyPrintf("service.email_queue", "[EmailQueue] Worker %d unknown task type: %s", workerID, task.TaskType)
	}
}

func (s *EmailQueueService) sendAnnouncementRecipients(announcementID int64, campaignID string, recipients []string, title, content string) (int, int, error) {
	if s == nil || s.emailService == nil || s.emailService.notificationEmailService == nil {
		return 0, len(recipients), ErrEmailNotConfigured
	}

	processed := 0
	failed := 0
	var lastErr error
	for _, recipient := range recipients {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := s.emailService.notificationEmailService.Send(ctx, NotificationEmailSendInput{
			Event:          NotificationEmailEventAnnouncementPublished,
			RecipientEmail: recipient,
			RecipientName:  emailRecipientName(recipient),
			SourceType:     "announcement",
			SourceID:       strconv.FormatInt(announcementID, 10),
			ReminderKey:    campaignID,
			Variables: map[string]string{
				"announcement_title":   title,
				"announcement_content": content,
			},
		})
		cancel()
		if err != nil {
			failed++
			lastErr = err
			logger.LegacyPrintf("service.email_queue", "[EmailQueue] Failed to send announcement to recipient %s: %v", notificationEmailHash(recipient), err)
			continue
		}
		processed++
	}
	if failed > 0 {
		return processed, failed, fmt.Errorf("%d announcement email recipients failed; last error: %w", failed, lastErr)
	}
	return processed, 0, nil
}

// EnqueueVerifyCode 将验证码发送任务加入队列
func (s *EmailQueueService) EnqueueVerifyCode(email, siteName string, locale ...string) error {
	task := EmailTask{
		Email:    email,
		SiteName: siteName,
		TaskType: TaskTypeVerifyCode,
		Locale:   firstEmailLocale(locale),
	}

	select {
	case s.taskChan <- task:
		logger.LegacyPrintf("service.email_queue", "[EmailQueue] Enqueued verify code task for %s", email)
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// EnqueuePasswordReset 将密码重置邮件任务加入队列
func (s *EmailQueueService) EnqueuePasswordReset(email, siteName, resetURL string, locale ...string) error {
	task := EmailTask{
		Email:    email,
		SiteName: siteName,
		TaskType: TaskTypePasswordReset,
		ResetURL: resetURL,
		Locale:   firstEmailLocale(locale),
	}

	select {
	case s.taskChan <- task:
		logger.LegacyPrintf("service.email_queue", "[EmailQueue] Enqueued password reset task for %s", email)
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// GetSMTPConfig validates that announcement email can be sent before the
// announcement is persisted and queued.
func (s *EmailQueueService) GetSMTPConfig(ctx context.Context) (*SMTPConfig, error) {
	if s == nil || s.emailService == nil {
		return nil, ErrEmailNotConfigured
	}
	return s.emailService.GetSMTPConfig(ctx)
}

func (s *EmailQueueService) WakeAnnouncementWorker() {
	if s == nil || s.announcementBatchRepo == nil {
		return
	}
	select {
	case s.announcementWake <- struct{}{}:
	default:
	}
}

func (s *EmailQueueService) ListAnnouncementBatches(ctx context.Context, announcementID int64) ([]AnnouncementEmailBatch, error) {
	if s == nil || s.announcementBatchRepo == nil {
		return nil, ErrEmailNotConfigured
	}
	return s.announcementBatchRepo.ListByAnnouncement(ctx, announcementID)
}

// Stop 停止队列服务
func (s *EmailQueueService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	logger.LegacyPrintf("service.email_queue", "%s", "[EmailQueue] All workers stopped")
}
