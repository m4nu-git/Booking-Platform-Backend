import { Queue } from 'bullmq';
import { getRedisConnObject } from '../config/redis.config';

export const HOTEL_INDEX_QUEUE = 'hotel-indexing-queue';

export const hotelIndexQueue = new Queue(HOTEL_INDEX_QUEUE, {
    connection: getRedisConnObject(),
});
