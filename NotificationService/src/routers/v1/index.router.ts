import express from 'express';
import pingRouter from './ping.router';
import emailRouter from './email.router';

const v1Router = express.Router();

v1Router.use('/ping',  pingRouter);
v1Router.use('/email', emailRouter);

export default v1Router;