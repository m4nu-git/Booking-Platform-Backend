import express from 'express';
import { validateRequestBody } from '../../validators';
import { createBookingSchema } from '../../validators/booking.validator';
import {
    cancelBookingHandler,
    confirmBookingHandler,
    createBookingHandler,
    getBookingByIdHandler,
    getBookingsByUserIdHandler,
} from '../../controllers/booking.controller';

const bookingRouter = express.Router();

// Filter / list routes must come before /:id to avoid "user" being captured as an id
bookingRouter.get('/user', getBookingsByUserIdHandler);
bookingRouter.get('/:id', getBookingByIdHandler);
bookingRouter.post('/:id/cancel', cancelBookingHandler);

bookingRouter.post('/', validateRequestBody(createBookingSchema), createBookingHandler);
bookingRouter.post('/confirm/:idempotencyKey', confirmBookingHandler);

export default bookingRouter;
