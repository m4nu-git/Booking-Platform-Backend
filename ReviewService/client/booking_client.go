package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type BookingData struct {
	Id      int64  `json:"id"`
	UserId  int64  `json:"userId"`
	HotelId int64  `json:"hotelId"`
	Status  string `json:"status"`
}

type BookingClient struct {
	BaseURL    string
	HttpClient *http.Client
}

func NewBookingClient(baseURL string) *BookingClient {
	return &BookingClient{
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
}

// GetBooking fetches a booking by ID from BookingService.
// Returns nil, nil when the booking is not found (404).
func (c *BookingClient) GetBooking(bookingId int64) (*BookingData, error) {
	url := fmt.Sprintf("%s/bookings/%d", c.BaseURL, bookingId)

	resp, err := c.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("booking service returned %d", resp.StatusCode)
	}

	var res struct {
		Data BookingData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode booking response: %w", err)
	}

	return &res.Data, nil
}
