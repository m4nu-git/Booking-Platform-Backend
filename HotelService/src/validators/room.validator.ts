import { z } from 'zod';

export const getAvailableRoomSchema = z.object({
    roomCategoryId: z.string({message: "roomCategoryId must be present"}),
    checkInDate: z.string({message: "checkInDate must be present"}),
    checkOutDate: z.string({message: "checkOutDate must be present"}),
});


export const updateBookingIdToRoomsSchema = z.object({
    roomIds : z.array(z.number()).nonempty({message:"roomIds cannot be empty"}),
    bookingId: z.number({message:"bookingId required"})
})