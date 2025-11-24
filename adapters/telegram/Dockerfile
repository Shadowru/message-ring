FROM node:18-alpine

WORKDIR /app

COPY package.json ./
RUN npm install

COPY src ./src

# Том для хранения сессий (обязательно должен совпадать с docker-compose)
VOLUME /app/store

CMD ["node", "src/index.js"]