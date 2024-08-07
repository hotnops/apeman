FROM node:20-alpine as build
WORKDIR /app
COPY package.json ./
RUN npm install && npm cache clean --force
COPY index.html ./
COPY public/ ./public/
COPY src/ ./src/
COPY tsconfig.json ./
COPY tsconfig.node.json ./
COPY vite.config.ts ./
RUN npm run build
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]