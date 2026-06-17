package insights

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	GetSummary(ctx context.Context, from, to *string) (*Summary, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetSummary(ctx context.Context, from, to *string) (*Summary, error) {
	filter := defaultMonthFilter()
	if from != nil && *from != "" {
		if _, err := time.Parse("2006-01-02", *from); err != nil {
			return nil, errors.New("invalid from date (expected format: YYYY-MM-DD)")
		}
		filter.From = *from
	}
	if to != nil && *to != "" {
		if _, err := time.Parse("2006-01-02", *to); err != nil {
			return nil, errors.New("invalid to date (expected format: YYYY-MM-DD)")
		}
		filter.To = *to
	}
	return s.repo.GetSummary(ctx, filter)
}

func defaultMonthFilter() SummaryFilter {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	end := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local)
	return SummaryFilter{
		From: start.Format("2006-01-02"),
		To:   end.Format("2006-01-02"),
	}
}
