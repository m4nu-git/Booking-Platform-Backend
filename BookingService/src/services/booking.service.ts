import { CreateBookingDTO } from "../dto/booking.dto";
import { confirmBooking, createBooking, createIdempotencyKey, finalizeIdempotencyKey, getIdempotencyKeyWithLock } from "../repositories/booking.repository";
import { BadRequestError, InternalServerError, NotFoundError } from "../utils/errors/app.error";
import { generateIdempotencyKey } from "../utils/generateIdempotencyKey";
import prismaClient from '../prisma/client';
import { getAvailableRooms, updateBookingIdToRooms } from "../api/hotel.api";
import { serverConfig } from "../config";
import { redlock } from "../config/redis.config";


type AvailableRoom = {
    id: number;
    roomCategoryId: number;
    dateOfAvailability: Date;
}


export async function createBookingService(
    createBookingDTO: CreateBookingDTO
) {
    const ttl = serverConfig.LOCK_TTL;
    const bookingResource = `hotel:${createBookingDTO.hotelId}`;


    const availableRoomsResponse = await getAvailableRooms(
        createBookingDTO.roomCategoryId,
        createBookingDTO.checkInDate,
        createBookingDTO.checkOutDate
    );

    const availableRooms = availableRoomsResponse.data || [];

    const checkInDate = new Date(createBookingDTO.checkInDate);
    const checkOutDate = new Date(createBookingDTO.checkOutDate);

    const totalNights = Math.ceil((checkOutDate.getTime() - checkInDate.getTime()) / (1000 * 60 * 60 * 24));

    if(availableRooms.length == 0 || availableRooms.length < totalNights) {
        throw new BadRequestError(`No rooms available for the given dates`)
    }

    try {
        await redlock.acquire([bookingResource], ttl);
        const booking = await createBooking({
        userId: createBookingDTO.userId,
        hotelId: createBookingDTO.hotelId,
        totalGuests: createBookingDTO.totalGuests,
        bookingAmount: createBookingDTO.bookingAmount,
        checkInDate: new Date(createBookingDTO.checkInDate),
        checkOutDate: new Date(createBookingDTO.checkOutDate),
        roomCategoryId: createBookingDTO.roomCategoryId
    });

    const idempotencyKey = generateIdempotencyKey();

    await createIdempotencyKey(idempotencyKey, booking.id);

    await updateBookingIdToRooms(booking.id, availableRooms.map((room: AvailableRoom) => room.id));

    return {
        bookingId: booking.id,
        idempotencyKey: idempotencyKey
    }
    } catch (error) {
        console.error('Booking creation error:', error);
        throw error instanceof BadRequestError ? error : new InternalServerError(`Failed to create booking: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
}


export async function confirmBookingService(idempotencyKey: string) {
    return await prismaClient.$transaction( async (tx) => {
        const idempotencyKeyData = await getIdempotencyKeyWithLock(tx, idempotencyKey);

        if(!idempotencyKeyData || !idempotencyKeyData.bookingId) {
            throw new NotFoundError('Idempotency key not found');
        }

        if(idempotencyKeyData.finalized) {
            throw new BadRequestError('Idempotency key already finalized');
        }

        const booking = await confirmBooking(tx, idempotencyKeyData.bookingId);
        await finalizeIdempotencyKey(tx, idempotencyKey);
        // todo: mark the rooms as null if booking is cancelled or failed!

        return booking;
    })
}