//go:build integration

package repository

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (s *RedeemCodeRepoSuite) TestListWithFilters_SortByValueAsc() {
	s.Require().NoError(s.repo.Create(s.ctx, &service.RedeemCode{Code: "VALUE-20", Type: service.RedeemTypeBalance, Value: 20, Status: service.StatusUnused}))
	s.Require().NoError(s.repo.Create(s.ctx, &service.RedeemCode{Code: "VALUE-10", Type: service.RedeemTypeBalance, Value: 10, Status: service.StatusUnused}))

	codes, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "value",
		SortOrder: "asc",
	}, "", "", "")
	s.Require().NoError(err)
	s.Require().Len(codes, 2)
	s.Require().Equal("VALUE-10", codes[0].Code)
	s.Require().Equal("VALUE-20", codes[1].Code)
}

func (s *RedeemCodeRepoSuite) TestListWithFilters_SortByActivityAtDesc() {
	user := s.createUser(uniqueTestValue(s.T(), "activity-sort") + "@example.com")
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	oldCode, err := s.client.RedeemCode.Create().
		SetCode("ACTIVITY-OLD-USED").
		SetType(service.RedeemTypeBalance).
		SetStatus(service.StatusUsed).
		SetValue(0).
		SetNotes("").
		SetValidityDays(30).
		SetCreatedAt(base).
		SetUsedBy(user.ID).
		SetUsedAt(base.Add(2 * time.Hour)).
		Save(s.ctx)
	s.Require().NoError(err)

	newCode, err := s.client.RedeemCode.Create().
		SetCode("ACTIVITY-NEW-UNUSED").
		SetType(service.RedeemTypeBalance).
		SetStatus(service.StatusUnused).
		SetValue(0).
		SetNotes("").
		SetValidityDays(30).
		SetCreatedAt(base.Add(time.Hour)).
		Save(s.ctx)
	s.Require().NoError(err)
	s.Require().Greater(newCode.ID, oldCode.ID)

	codes, _, err := s.repo.ListWithFilters(s.ctx, pagination.PaginationParams{
		Page:      1,
		PageSize:  10,
		SortBy:    "activity_at",
		SortOrder: "desc",
	}, "", "", "")
	s.Require().NoError(err)
	s.Require().Len(codes, 2)
	s.Require().Equal("ACTIVITY-OLD-USED", codes[0].Code)
	s.Require().Equal("ACTIVITY-NEW-UNUSED", codes[1].Code)
}
