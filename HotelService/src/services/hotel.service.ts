import { createHotelDTO, updateHotelDto } from "../dto/hotel.dto";
import { HotelRepository } from "../repositories/hotel.repository";
import { addHotelDeleteJob, addHotelIndexJob } from "../producers/hotelIndex.producer";
import logger from "../config/logger.config";

const hotelRepository = new HotelRepository();

// Queues an ES indexing job without blocking the HTTP response.
// If the queue or Redis has a transient issue the hotel mutation still succeeds
// and the next successful queue call will pick it up.
async function scheduleIndex(hotelId: number, action: 'index' | 'delete') {
    try {
        if (action === 'index') {
            await addHotelIndexJob(hotelId);
        } else {
            await addHotelDeleteJob(hotelId);
        }
    } catch (err) {
        logger.error(`Failed to queue ES ${action} job for hotel ${hotelId}: ${err}`);
    }
}

export async function createHotelService(hotelData: createHotelDTO) {
    const hotel = await hotelRepository.create(hotelData);
    scheduleIndex(hotel.id, 'index');   // intentionally not awaited — fire and forget
    return hotel;
}

export async function getHotelByIdService(id: number) {
    return hotelRepository.findById(id);
}

export async function getAllHotelsService() {
    return hotelRepository.findAll();
}

export async function updateHotelService(id: number, hotelData: updateHotelDto) {
    const updated = await hotelRepository.update(id, hotelData);
    scheduleIndex(id, 'index');         // intentionally not awaited — fire and forget
    return updated;
}

export async function deleteHotelService(id: number) {
    const result = await hotelRepository.softDelete(id);
    scheduleIndex(id, 'delete');        // intentionally not awaited — fire and forget
    return result;
}
