import esClient, { ES_HOTEL_INDEX } from '../config/elasticsearch.config';
import logger from '../config/logger.config';

export class ElasticsearchRepository {

    async ensureIndex(): Promise<void> {
        const exists = await esClient.indices.exists({ index: ES_HOTEL_INDEX });
        if (!exists) {
            await esClient.indices.create({
                index: ES_HOTEL_INDEX,
                mappings: {
                    properties: {
                        id:          { type: 'integer' },
                        name:        { type: 'text',    analyzer: 'standard' },
                        address:     { type: 'text',    analyzer: 'standard' },
                        location:    { type: 'keyword' },
                        rating:      { type: 'float' },
                        ratingCount: { type: 'integer' },
                        minPrice:    { type: 'integer' },
                        roomTypes:   { type: 'keyword' },
                    },
                },
            });
            logger.info(`Created Elasticsearch index: ${ES_HOTEL_INDEX}`);
        }
    }

    async indexDocument(id: string, doc: object): Promise<void> {
        await esClient.index({ index: ES_HOTEL_INDEX, id, document: doc });
        logger.info(`Indexed hotel ${id} in Elasticsearch`);
    }

    async deleteDocument(id: string): Promise<void> {
        try {
            await esClient.delete({ index: ES_HOTEL_INDEX, id });
            logger.info(`Deleted hotel ${id} from Elasticsearch`);
        } catch (err: any) {
            if (err?.meta?.statusCode !== 404) throw err;
        }
    }

    async search(query: object, from = 0, size = 10): Promise<any[]> {
        const result = await esClient.search({ index: ES_HOTEL_INDEX, ...query, from, size });
        return (result.hits.hits as any[]).map((h: any) => ({ ...h._source, _score: h._score }));
    }
}
