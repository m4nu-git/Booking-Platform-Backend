import { Request, Response, NextFunction } from 'express';
import { searchHotelsService } from '../services/search.service';
import { StatusCodes } from 'http-status-codes';

export async function searchHotelsHandler(req: Request, res: Response, next: NextFunction) {
    try {
        const results = await searchHotelsService({
            name:      req.query.name      as string | undefined,
            location:  req.query.location  as string | undefined,
            minRating: req.query.minRating ? Number(req.query.minRating) : undefined,
            maxRating: req.query.maxRating ? Number(req.query.maxRating) : undefined,
            roomType:  req.query.roomType  as string | undefined,
            minPrice:  req.query.minPrice  ? Number(req.query.minPrice)  : undefined,
            maxPrice:  req.query.maxPrice  ? Number(req.query.maxPrice)  : undefined,
            page:      req.query.page      ? Number(req.query.page)      : 1,
            limit:     req.query.limit     ? Number(req.query.limit)     : 10,
        });

        res.status(StatusCodes.OK).json({
            success: true,
            message: 'Search results',
            data:    results,
        });
    } catch (err: any) {
        res.status(StatusCodes.SERVICE_UNAVAILABLE).json({
            success: false,
            message: err.message,
        });
    }
}
