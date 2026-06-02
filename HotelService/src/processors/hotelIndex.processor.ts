import { Job, Worker } from 'bullmq';
import { HOTEL_INDEX_QUEUE } from '../queues/hotelIndex.queue';
import { getRedisConnObject } from '../config/redis.config';
import { ElasticsearchRepository } from '../elasticsearch/elasticsearch.repository';
import { transformHotelToESDoc } from '../elasticsearch/hotel.transformer';
import Hotel from '../db/models/hotel';
import RoomCategory from '../db/models/roomCategory';
import logger from '../config/logger.config';

const esRepository = new ElasticsearchRepository();

// Directly indexes a single hotel (bypasses BullMQ — used for bulk sync at startup)
async function indexHotelNow(hotelId: number): Promise<void> {
    const hotel = await Hotel.findByPk(hotelId);
    if (!hotel) return;

    const roomCategories = await RoomCategory.findAll({
        where: { hotelId, deletedAt: null },
    });

    const doc = transformHotelToESDoc(hotel, roomCategories);
    await esRepository.indexDocument(String(hotelId), doc);
}

// Fetches every active hotel from MySQL and indexes it into ES.
// Runs once at startup so that hotels created before ES was available appear in search.
async function syncAllHotelsToES(): Promise<void> {
    const hotels = await Hotel.findAll({ where: { deletedAt: null } });

    if (hotels.length === 0) {
        logger.info('No hotels to sync into Elasticsearch');
        return;
    }

    logger.info(`Syncing ${hotels.length} hotel(s) into Elasticsearch...`);

    let indexed = 0;
    for (const hotel of hotels) {
        try {
            await indexHotelNow(hotel.id);
            indexed++;
        } catch (err) {
            logger.error(`Failed to index hotel ${hotel.id} during startup sync: ${err}`);
        }
    }

    logger.info(`Elasticsearch startup sync complete — ${indexed}/${hotels.length} hotel(s) indexed`);
}

export async function setupHotelIndexWorker(): Promise<void> {
    // Try to ensure the index exists, but don't crash the server if ES is not running.
    // Individual jobs will fail + be retried once ES becomes available.
    try {
        await esRepository.ensureIndex();
        logger.info(`Elasticsearch index "${process.env.ES_HOTEL_INDEX || 'hotels'}" is ready`);

        // Index all existing hotels so that hotels created before ES was running
        // appear in search results immediately.
        await syncAllHotelsToES();
    } catch (err) {
        logger.warn(`Elasticsearch not reachable at startup — hotel indexing jobs will be retried when ES is available. (${err})`);
    }

    const worker = new Worker(
        HOTEL_INDEX_QUEUE,
        async (job: Job) => {
            const { action, hotelId } = job.data;

            if (action === 'delete') {
                await esRepository.deleteDocument(String(hotelId));
                return;
            }

            // action === 'index': fetch hotel + room categories and index them
            const hotel = await Hotel.findByPk(hotelId);
            if (!hotel) {
                logger.warn(`Hotel ${hotelId} not found — skipping ES index`);
                return;
            }

            const roomCategories = await RoomCategory.findAll({
                where: { hotelId, deletedAt: null },
            });

            const doc = transformHotelToESDoc(hotel, roomCategories);
            await esRepository.indexDocument(String(hotelId), doc);
        },
        { connection: getRedisConnObject() },
    );

    worker.on('failed', (job, err) => {
        logger.error(`Hotel index job failed [jobId=${job?.id}]: ${err.message}`);
    });

    logger.info('Hotel Elasticsearch indexing worker started');
}
