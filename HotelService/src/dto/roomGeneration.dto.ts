import { z } from 'zod';

export const RoomgenerationRequestSchema = {
    roomCategoryId: z.number().positive(),
    startDate: z.string().datetime(),
    endDate: z.string().datetime(),
    scheduleType: z.enum(['immediate', 'scheduled']).default('immediate'),
    scheduleAt: z.string().datetime().optional(),
    priceOverride: z.number().positive().optional()
}