import Hotel from '../db/models/hotel';
import RoomCategory from '../db/models/roomCategory';

export function transformHotelToESDoc(hotel: Hotel, roomCategories: RoomCategory[]) {
    const prices = roomCategories.map(rc => rc.price).filter(p => p != null);
    const roomTypes = [...new Set(roomCategories.map(rc => rc.roomType))];

    return {
        id:          hotel.id,
        name:        hotel.name,
        address:     hotel.address,
        location:    hotel.location,
        rating:      hotel.rating     ?? 0,
        ratingCount: hotel.ratingCount ?? 0,
        minPrice:    prices.length ? Math.min(...prices) : 0,
        roomTypes,
    };
}
