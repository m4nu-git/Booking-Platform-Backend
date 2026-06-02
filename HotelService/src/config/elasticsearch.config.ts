import { Client } from '@elastic/elasticsearch';

const esClient = new Client({
    node: process.env.ES_NODE || 'http://localhost:9200',
    auth: process.env.ES_USERNAME
        ? { username: process.env.ES_USERNAME, password: process.env.ES_PASSWORD || '' }
        : undefined,
    tls: process.env.ES_NODE?.startsWith('https') ? { rejectUnauthorized: false } : undefined,
    requestTimeout: 3000,   // fail fast — don't hang for 30 s when ES is down
    maxRetries: 1,
});

export const ES_HOTEL_INDEX = process.env.ES_HOTEL_INDEX || 'hotels';

export default esClient;
