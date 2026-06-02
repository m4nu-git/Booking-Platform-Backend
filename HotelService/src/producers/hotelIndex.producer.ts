import { hotelIndexQueue } from '../queues/hotelIndex.queue';

export async function addHotelIndexJob(hotelId: number): Promise<void> {
    await hotelIndexQueue.add('hotel-index', { action: 'index', hotelId });
}

export async function addHotelDeleteJob(hotelId: number): Promise<void> {
    await hotelIndexQueue.add('hotel-index', { action: 'delete', hotelId });
}
