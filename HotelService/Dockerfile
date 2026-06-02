FROM node:20-alpine

WORKDIR /app

# Install dependencies (separate layer for better caching)
COPY package*.json ./
RUN npm ci

COPY . .

EXPOSE 3000

# Run Sequelize migrations then start the server
CMD ["sh", "-c", "npm run migrate && npm start"]
