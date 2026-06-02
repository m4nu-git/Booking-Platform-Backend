FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

# Generate the Prisma client — schema lives at src/prisma/schema.prisma (non-default location)
RUN npx prisma generate --schema src/prisma/schema.prisma

EXPOSE 3001

# Run Prisma migrations then start the server
CMD ["sh", "-c", "npx prisma migrate deploy --schema src/prisma/schema.prisma && npm start"]
