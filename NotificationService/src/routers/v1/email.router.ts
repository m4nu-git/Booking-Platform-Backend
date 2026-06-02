import express from 'express';
import { sendEmailHandler } from '../../controllers/email.controller';

const emailRouter = express.Router();

emailRouter.post('/send', sendEmailHandler);

export default emailRouter;
