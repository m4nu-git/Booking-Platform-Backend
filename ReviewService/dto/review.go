package dto

type CreateReviewRequestDTO struct {
	UserId    int64  `json:"user_id" validate:"required"`
	BookingId int64  `json:"booking_id" validate:"required"`
	HotelId   int64  `json:"hotel_id" validate:"required"`
	Comment   string `json:"comment" validate:"required,min=1,max=1000"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
}
