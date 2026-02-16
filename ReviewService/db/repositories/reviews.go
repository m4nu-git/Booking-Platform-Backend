package db

import (
	"ReviewService/models"
	"database/sql"
	"fmt"
)

type ReviewRepository interface {
	Create(userId int64, bookingId int64, hotelId int64, comment string, rating int) (*models.Review, error)
}

type ReviewRepositoryImpl struct {
	db *sql.DB
}

func NewReviewRepository(_db *sql.DB) ReviewRepository {
	return &ReviewRepositoryImpl{
		db: _db,
	}
}

func (r *ReviewRepositoryImpl) Create(userId int64, bookingId int64, hotelId int64, comment string, rating int) (*models.Review, error) {
	query := "INSERT INTO reviews (user_id, booking_id, hotel_id, comment, rating) VALUES (?, ?, ?, ?, ?)"
	result, err := r.db.Exec(query, userId, bookingId, hotelId, comment, rating)

	if err != nil {
		fmt.Println("Error creating review:", err)
		return nil, err
	}

	lastInsertID, rowErr := result.LastInsertId()
	if rowErr != nil {
		fmt.Println("Error getting last insert ID:", rowErr)
		return nil, rowErr
	}

	review := &models.Review{
		Id:        lastInsertID,
		UserId:    userId,
		BookingId: bookingId,
		HotelId:   hotelId,
		Comment:   comment,
		Rating:    rating,
		IsSynced:  false,
	}

	fmt.Println("Review created successfully:", review)
	return review, nil
}
