import { ElasticsearchRepository } from '../elasticsearch/elasticsearch.repository';

const esRepository = new ElasticsearchRepository();

export interface HotelSearchFiltersDTO {
    name?:      string;
    location?:  string;
    minRating?: number;
    maxRating?: number;
    roomType?:  string;
    minPrice?:  number;
    maxPrice?:  number;
    page?:      number;
    limit?:     number;
}

export async function searchHotelsService(filters: HotelSearchFiltersDTO) {
    const {
        name, location,
        minRating, maxRating,
        roomType,
        minPrice, maxPrice,
        page  = 1,
        limit = 10,
    } = filters;

    const must: object[] = name
        ? [{ multi_match: { query: name, fields: ['name^3', 'address^2'], fuzziness: 'AUTO' } }]
        : [{ match_all: {} }];

    const filter: object[] = [];

    if (location) filter.push({ term: { location } });

    if (minRating != null || maxRating != null) {
        const range: Record<string, number> = {};
        if (minRating != null) range.gte = minRating;
        if (maxRating != null) range.lte = maxRating;
        filter.push({ range: { rating: range } });
    }

    if (minPrice != null || maxPrice != null) {
        const range: Record<string, number> = {};
        if (minPrice != null) range.gte = minPrice;
        if (maxPrice != null) range.lte = maxPrice;
        filter.push({ range: { minPrice: range } });
    }

    if (roomType) filter.push({ term: { roomTypes: roomType } });

    const query = {
        query: { bool: { must, filter } },
        sort:  [{ _score: 'desc' }, { rating: 'desc' }],
    };

    const from = (page - 1) * Math.min(limit, 50);

    try {
        return await esRepository.search(query, from, Math.min(limit, 50));
    } catch (err: any) {
        // Surface a readable error so the caller knows ES is down, not the app
        const msg = err?.message ?? String(err);
        throw new Error(`Elasticsearch is not available — start ES at ${process.env.ES_NODE || 'http://localhost:9200'} to use hotel search. (${msg})`);
    }
}
