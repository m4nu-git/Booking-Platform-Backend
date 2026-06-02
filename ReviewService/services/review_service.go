package services

import (
	"ReviewService/client"
	db "ReviewService/db/repositories"
	"ReviewService/dto"
	"ReviewService/models"
	"ReviewService/utils"
	"fmt"
	"net/http"
	"strconv"
)

type ReviewService interface {
	CreateReview(payload *dto.CreateReviewRequestDTO) (*models.Review, error)
	GetReviewById(id string) (*models.Review, error)
	UpdateReview(id string, payload *dto.UpdateReviewRequestDTO) (*models.Review, error)
	DeleteReview(id string) error
	GetAllReviews() ([]*models.Review, error)
	GetReviewsByUserId(userId string) ([]*models.Review, error)
	GetReviewsByHotelId(hotelId string) ([]*models.Review, error)
	GetReviewsByBookingId(bookingId string) ([]*models.Review, error)
}

type ReviewServiceImpl struct {
	reviewRepository db.ReviewRepository
	bookingClient    *client.BookingClient
}

func NewReviewService(_reviewRepository db.ReviewRepository, _bookingClient *client.BookingClient) ReviewService {
	return &ReviewServiceImpl{
		reviewRepository: _reviewRepository,
		bookingClient:    _bookingClient,
	}
}

func (r *ReviewServiceImpl) CreateReview(payload *dto.CreateReviewRequestDTO) (*models.Review, error) {
	fmt.Println("Creating review in ReviewService")

	if payload.Rating < 1 || payload.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	// Validate booking ownership before writing the review.
	booking, err := r.bookingClient.GetBooking(payload.BookingId)
	if err != nil {
		return nil, fmt.Errorf("could not verify booking: %w", err)
	}
	if booking == nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "booking not found")
	}
	if booking.UserId != payload.UserId {
		return nil, utils.NewServiceError(http.StatusForbidden, "booking does not belong to this user")
	}
	if booking.HotelId != payload.HotelId {
		return nil, utils.NewServiceError(http.StatusBadRequest, "booking hotel does not match review hotel")
	}
	if booking.Status != "CONFIRMED" {
		return nil, utils.NewServiceError(http.StatusBadRequest, "can only review a confirmed booking")
	}

	review, err := r.reviewRepository.Create(payload.UserId, payload.BookingId, payload.HotelId, payload.Comment, payload.Rating)
	if err != nil {
		fmt.Println("Error creating review:", err)
		return nil, err
	}

	fmt.Println("Review created successfully:", review)
	return review, nil
}

func (r *ReviewServiceImpl) GetReviewById(id string) (*models.Review, error) {
	fmt.Println("Fetching review in ReviewService")

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid review ID")
	}

	return r.reviewRepository.GetByID(idInt)
}

func (r *ReviewServiceImpl) UpdateReview(id string, payload *dto.UpdateReviewRequestDTO) (*models.Review, error) {
	fmt.Println("Updating review in ReviewService")

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid review ID")
	}

	if payload.Rating < 1 || payload.Rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	return r.reviewRepository.Update(idInt, payload.Comment, payload.Rating)
}

func (r *ReviewServiceImpl) DeleteReview(id string) error {
	fmt.Println("Deleting review in ReviewService")

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid review ID")
	}

	return r.reviewRepository.Delete(idInt)
}

func (r *ReviewServiceImpl) GetAllReviews() ([]*models.Review, error) {
	return r.reviewRepository.GetAll()
}

func (r *ReviewServiceImpl) GetReviewsByUserId(userId string) ([]*models.Review, error) {
	userIdInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}
	return r.reviewRepository.GetByUserId(userIdInt)
}

func (r *ReviewServiceImpl) GetReviewsByHotelId(hotelId string) ([]*models.Review, error) {
	hotelIdInt, err := strconv.ParseInt(hotelId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid hotel ID")
	}
	return r.reviewRepository.GetByHotelId(hotelIdInt)
}

func (r *ReviewServiceImpl) GetReviewsByBookingId(bookingId string) ([]*models.Review, error) {
	bookingIdInt, err := strconv.ParseInt(bookingId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid booking ID")
	}
	return r.reviewRepository.GetByBookingId(bookingIdInt)
}
