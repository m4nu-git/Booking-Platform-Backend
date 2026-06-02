import { CreateBookingDTO } from "../dto/booking.dto";
import { cancelBooking, confirmBooking, createBooking, createIdempotencyKey, finalizeIdempotencyKey, getBookingById, getBookingsByUserId, getIdempotencyKeyWithLock } from "../repositories/booking.repository";
import { BadRequestError, InternalServerError, NotFoundError } from "../utils/errors/app.error";
import { generateIdempotencyKey } from "../utils/generateIdempotencyKey";
import prismaClient from '../prisma/client';
import { getAvailableRooms, releaseRoomsForBooking, updateBookingIdToRooms } from "../api/hotel.api";
import { serverConfig } from "../config";
import { redlock } from "../config/redis.config";
import { addEmailToQueue } from "../producers/email.producer";


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


export async function getBookingByIdService(bookingId: number) {
    const booking = await getBookingById(bookingId);
    if (!booking) throw new NotFoundError('Booking not found');
    return booking;
}

export async function getBookingsByUserIdService(userId: number) {
    return getBookingsByUserId(userId);
}

export async function cancelBookingService(bookingId: number, userEmail?: string) {
    const booking = await getBookingById(bookingId);
    if (!booking) throw new NotFoundError('Booking not found');
    if (booking.status === 'CANCELLED') throw new BadRequestError('Booking is already cancelled');

    await releaseRoomsForBooking(bookingId);
    const cancelled = await cancelBooking(bookingId);

    if (userEmail) {
        await addEmailToQueue({
            to: userEmail,
            subject: 'Your Booking Has Been Cancelled',
            templateId: 'booking-cancelled',
            params: {
                name: userEmail.split('@')[0],
                bookingId,
                cancellationDate: new Date().toLocaleDateString(),
            },
        });
    }

    return cancelled;
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