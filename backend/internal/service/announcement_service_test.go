package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type announcementRepoStub struct {
	item *Announcement
}

type announcementEmailUserRepoStub struct {
	UserRepository
	users                []User
	includeSubscriptions *bool
}

func (s *announcementEmailUserRepoStub) ListWithFilters(_ context.Context, params pagination.PaginationParams, filters UserListFilters) ([]User, *pagination.PaginationResult, error) {
	s.includeSubscriptions = filters.IncludeSubscriptions
	return s.users, &pagination.PaginationResult{
		Total:    int64(len(s.users)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    1,
	}, nil
}

func (s *announcementRepoStub) Create(_ context.Context, a *Announcement) error {
	s.item = a
	return nil
}

func (s *announcementRepoStub) CreateWithEmailBatch(_ context.Context, a *Announcement, _ *AnnouncementEmailBatch) error {
	return s.Create(context.Background(), a)
}

func (s *announcementRepoStub) GetByID(_ context.Context, _ int64) (*Announcement, error) {
	if s.item == nil {
		return nil, ErrAnnouncementNotFound
	}
	return s.item, nil
}

func (s *announcementRepoStub) Update(_ context.Context, a *Announcement) error {
	s.item = a
	return nil
}

func (s *announcementRepoStub) UpdateWithEmailBatch(_ context.Context, a *Announcement, _ *AnnouncementEmailBatch) error {
	return s.Update(context.Background(), a)
}

func (*announcementRepoStub) Delete(context.Context, int64) error {
	return nil
}

func (*announcementRepoStub) List(context.Context, pagination.PaginationParams, AnnouncementListFilters) ([]Announcement, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (*announcementRepoStub) ListActive(context.Context, time.Time) ([]Announcement, error) {
	return nil, nil
}

func TestAnnouncementServiceCreateRejectsEqualStartEndTimes(t *testing.T) {
	repo := &announcementRepoStub{}
	svc := NewAnnouncementService(repo, nil, nil, nil, nil)
	now := time.Unix(1776790020, 0)

	_, err := svc.Create(context.Background(), &CreateAnnouncementInput{
		Title:      "公告",
		Content:    "内容",
		Status:     AnnouncementStatusActive,
		NotifyMode: AnnouncementNotifyModePopup,
		StartsAt:   &now,
		EndsAt:     &now,
	})
	require.ErrorIs(t, err, ErrAnnouncementInvalidSchedule)
}

func TestAnnouncementServiceUpdateRejectsEqualStartEndTimes(t *testing.T) {
	repo := &announcementRepoStub{
		item: &Announcement{
			ID:         1,
			Title:      "公告",
			Content:    "内容",
			Status:     AnnouncementStatusActive,
			NotifyMode: AnnouncementNotifyModePopup,
		},
	}
	svc := NewAnnouncementService(repo, nil, nil, nil, nil)
	now := time.Unix(1776790020, 0)
	startsAt := &now
	endsAt := &now

	_, err := svc.Update(context.Background(), 1, &UpdateAnnouncementInput{
		StartsAt: &startsAt,
		EndsAt:   &endsAt,
	})
	require.ErrorIs(t, err, ErrAnnouncementInvalidSchedule)
}

func TestAnnouncementServiceEmailRecipientsRespectTargetingAndDeduplicate(t *testing.T) {
	now := time.Now()
	userRepo := &announcementEmailUserRepoStub{users: []User{
		{ID: 1, Email: "Eligible@example.com", Status: StatusActive, Balance: 20},
		{ID: 2, Email: "eligible@example.com", Status: StatusActive, Balance: 30},
		{ID: 3, Email: "low@example.com", Status: StatusActive, Balance: 2},
		{ID: 4, Email: "sub@example.com", Status: StatusActive, Balance: 0, Subscriptions: []UserSubscription{
			{GroupID: 9, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
		}},
	}}
	svc := NewAnnouncementService(nil, nil, userRepo, nil, nil)
	targeting := AnnouncementTargeting{AnyOf: []AnnouncementConditionGroup{
		{AllOf: []AnnouncementCondition{{Type: AnnouncementConditionTypeBalance, Operator: AnnouncementOperatorGT, Value: 10}}},
		{AllOf: []AnnouncementCondition{{Type: AnnouncementConditionTypeSubscription, Operator: AnnouncementOperatorIn, GroupIDs: []int64{9}}}},
	}}

	recipients, err := svc.listAnnouncementEmailRecipients(context.Background(), targeting)

	require.NoError(t, err)
	require.Equal(t, []string{"Eligible@example.com", "sub@example.com"}, recipients)
	require.NotNil(t, userRepo.includeSubscriptions)
	require.True(t, *userRepo.includeSubscriptions)
}

func TestAnnouncementServiceEmailRequiresActiveStatus(t *testing.T) {
	svc := NewAnnouncementService(&announcementRepoStub{}, nil, nil, nil, nil)

	_, err := svc.Create(context.Background(), &CreateAnnouncementInput{
		Title: "公告", Content: "内容", Status: AnnouncementStatusDraft, SendEmail: true,
	})

	require.ErrorIs(t, err, ErrAnnouncementEmailRequiresActive)
}
