import { Request, Response } from "express";
import { getAvailableRoomsService, updateBookingIdToRoomsService } from "../services/room.service";
import { StatusCodes } from "http-status-codes";


export async function getAvailableRoomsController(req: Request, res: Response) {
    
    const availableRooms = await getAvailableRoomsService({
        roomCategoryId: Number(req.query.roomCategoryId),
        checkInDate: new Date(req.query.checkInDate as string),
        checkOutDate: new Date(req.query.checkOutDate as string)
    });
    res.status(StatusCodes.OK).json({
        success: true,
        message: "Available rooms fetched successfully",
        data: availableRooms
    });
}

export async function updateBookingIdToRoomsController(req: Request, res: Response) {
    const result = await updateBookingIdToRoomsService(req.body);
    res.status(StatusCodes.OK).json({
        success: true,
        message: "Booking IDs updated to rooms table successfully",
        data: result
    });
}