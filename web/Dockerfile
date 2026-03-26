# web/Dockerfile
# Stage 1: build the React app
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# Stage 2: serve with Nginx
FROM nginx:1.27-alpine

# Remove default config
RUN rm /etc/nginx/conf.d/default.conf

# Our custom config (proxies /api/ to the Go service)
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the Vite build output
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
