package services

import (
	db "ReviewService/db/repositories"
	"ReviewService/dto"
	"ReviewService/models"
	"fmt"
)

type ReviewService interface {
	CreateReview(payload *dto.CreateReviewRequestDTO) (*models.Review, error)
}

type ReviewServiceImpl struct {
	reviewRepository db.ReviewRepository
}

func NewReviewService(_reviewRepository db.ReviewRepository) ReviewService {
	return &ReviewServiceImpl{
		reviewRepository: _reviewRepository,
	}
}

func (r *ReviewServiceImpl) CreateReview(payload *dto.CreateReviewRequestDTO) (*models.Review, error) {
	fmt.Println("Creating review in ReviewService")

	// Validate rating range
	if payload.Rating < 1 || payload.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	// Call the repository to create the review
	review, err := r.reviewRepository.Create(payload.UserId, payload.BookingId, payload.HotelId, payload.Comment, payload.Rating)
	if err != nil {
		fmt.Println("Error creating review:", err)
		return nil, err
	}

	fmt.Println("Review created successfully:", review)
	return review, nil
}
