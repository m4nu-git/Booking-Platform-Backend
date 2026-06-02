FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

EXPOSE 3002

# No migrations needed — NotificationService has no database
CMD ["npm", "start"]
