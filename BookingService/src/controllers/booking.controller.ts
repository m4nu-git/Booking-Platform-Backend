import { Request, Response } from "express";
import {
    cancelBookingService,
    confirmBookingService,
    createBookingService,
    getBookingByIdService,
    getBookingsByUserIdService,
} from "../services/booking.service";

export const createBookingHandler = async (req: Request, res: Response) => {
    const booking = await createBookingService(req.body);
    res.status(201).json({
        bookingId: booking.bookingId,
        idempotencyKey: booking.idempotencyKey,
    });
};

export const confirmBookingHandler = async (req: Request, res: Response) => {
    const booking = await confirmBookingService(req.params.idempotencyKey);
    res.status(200).json({
        bookingId: booking.id,
        status: booking.status,
    });
};

export const getBookingByIdHandler = async (req: Request, res: Response) => {
    const booking = await getBookingByIdService(Number(req.params.id));
    res.status(200).json({ success: true, message: "Booking fetched successfully", data: booking });
};

export const getBookingsByUserIdHandler = async (req: Request, res: Response) => {
    const userId = Number(req.query.userId);
    if (!userId) {
        res.status(400).json({ success: false, message: "userId query param is required" });
        return;
    }
    const bookings = await getBookingsByUserIdService(userId);
    res.status(200).json({ success: true, message: "Bookings fetched successfully", data: bookings });
};

export const cancelBookingHandler = async (req: Request, res: Response) => {
    const bookingId = Number(req.params.id);
    const userEmail = req.headers['x-user-email'] as string | undefined;
    const booking = await cancelBookingService(bookingId, userEmail);
    res.status(200).json({ success: true, message: "Booking cancelled successfully", data: booking });
};
