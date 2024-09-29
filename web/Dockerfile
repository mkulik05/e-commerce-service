FROM node:22-alpine

WORKDIR /statics

COPY package.json package-lock.json ./
RUN npm install --verbose

COPY . .

CMD ["npm", "run", "build"]