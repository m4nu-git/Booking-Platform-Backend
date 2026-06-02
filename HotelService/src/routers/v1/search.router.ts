import express from 'express';
import { searchHotelsHandler } from '../../controllers/search.controller';

const searchRouter = express.Router();

searchRouter.get('/search', searchHotelsHandler);

export default searchRouter;
