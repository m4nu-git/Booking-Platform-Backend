import { Request, Response } from 'express';
import { renderMailTemplate } from '../templates/templates.handler';
import { sendEmail } from '../services/mailer.service';
import { BadRequestError } from '../utils/errors/app.error';

// Direct synchronous email send — used by services that cannot push to BullMQ
// (e.g. AuthService calling over HTTP for password reset emails).
export async function sendEmailHandler(req: Request, res: Response) {
    const { to, subject, templateId, params } = req.body;

    if (!to || !subject || !templateId) {
        throw new BadRequestError('to, subject, and templateId are required');
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(to)) {
        throw new BadRequestError(`Invalid email address: "${to}". Expected format: user@domain.com`);
    }

    const html = await renderMailTemplate(templateId, params ?? {});
    await sendEmail(to, subject, html);

    res.status(200).json({ success: true, message: 'Email sent successfully' });
}
