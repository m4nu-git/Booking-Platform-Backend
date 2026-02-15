import express from "express";
import { validateQueryParams, validateRequestBody } from "../../validators";
import { getAvailableRoomSchema, updateBookingIdToRoomsSchema } from "../../validators/room.validator";
import { getAvailableRoomsController, updateBookingIdToRoomsController } from "../../controllers/room.controller";

const roomRouter = express.Router();

roomRouter.get("/available", validateQueryParams(getAvailableRoomSchema), getAvailableRoomsController);
roomRouter.put("/update-booking-ids", validateRequestBody(updateBookingIdToRoomsSchema), updateBookingIdToRoomsController);

export default roomRouter;