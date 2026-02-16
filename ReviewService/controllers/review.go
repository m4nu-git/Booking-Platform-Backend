package controllers

import (
	"ReviewService/dto"
	"ReviewService/services"
	"ReviewService/utils"
	"fmt"
	"net/http"
)

type ReviewController struct {
	ReviewService services.ReviewService
}

func NewReviewController(_reviewService services.ReviewService) *ReviewController {
	return &ReviewController{
		ReviewService: _reviewService,
	}
}

func (rc *ReviewController) CreateReview(w http.ResponseWriter, r *http.Request) {
	payload := r.Context().Value("payload").(dto.CreateReviewRequestDTO)

	fmt.Println("Payload received:", payload)

	review, err := rc.ReviewService.CreateReview(&payload)

	if err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "Failed to create review", err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "Review created successfully", review)
	fmt.Println("Review created successfully:", review)
}
