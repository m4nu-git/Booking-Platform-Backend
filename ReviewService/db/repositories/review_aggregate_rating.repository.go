package db

import (
	"context"
	"database/sql"
	"fmt"
)

type aggregateRating struct {
	HotelID int64
	Sum     float64
	Count   int64
}

type ReviewAggregateRatingRepository interface {
	FetchUnappliedAggregates(ctx context.Context, cutoff string) ([]aggregateRating, error)
	MarkReviewsAsSynced(ctx context.Context, tx *sql.Tx, HotelID int64, cutoff string) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type ReviewAggregateRatingRepositoryImpl struct {
	db *sql.DB
}

func NewReviewAggregateRatingRepository(_db *sql.DB) ReviewAggregateRatingRepository {
	return &ReviewAggregateRatingRepositoryImpl{
		db: _db,
	}
}

func (r *ReviewAggregateRatingRepositoryImpl) FetchUnappliedAggregates(ctx context.Context, cutoff string) ([]aggregateRating, error) {
	query := `
		SELECT hotel_id, SUM(rating) AS total_rating, COUNT(*) AS cnt
		FROM reviews
		WHERE is_synced = FALSE AND created_at <= ?
		GROUP BY hotel_id
	`

	rows, err := r.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query aggregates: %w", err)
	}
	defer rows.Close()

	var results []aggregateRating
	for rows.Next() {
		var a aggregateRating
		var sum sql.NullFloat64
		var cnt sql.NullInt64

		if err := rows.Scan(&a.HotelID, &sum, &cnt); err != nil {
			return nil, fmt.Errorf("scan agg row: %w", err)
		}
		if sum.Valid {
			a.Sum = sum.Float64
		}
		if cnt.Valid {
			a.Count = cnt.Int64
		}
		results = append(results, a)
	}
	return results, rows.Err()
}

func (r *ReviewAggregateRatingRepositoryImpl) MarkReviewsAsSynced(ctx context.Context, tx *sql.Tx, hotelID int64, cutoff string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE reviews 
		SET is_synced = TRUE 
		WHERE is_synced = FALSE AND hotel_id = ? AND created_at <= ?`, hotelID, cutoff)
	return err
}

func (r *ReviewAggregateRatingRepositoryImpl) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
