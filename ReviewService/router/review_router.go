package router

import (
	"ReviewService/controllers"
	"ReviewService/middlewares"

	"github.com/go-chi/chi/v5"
)

type ReviewRouter struct {
	reviewController *controllers.ReviewController
}

func NewReviewRouter(_reviewController *controllers.ReviewController) Router {
	return &ReviewRouter{
		reviewController: _reviewController,
	}
}

func (rr *ReviewRouter) Register(r chi.Router) {
	r.With(middlewares.ReviewCreateRequestValidator).Post("/reviews", rr.reviewController.CreateReview)
}
