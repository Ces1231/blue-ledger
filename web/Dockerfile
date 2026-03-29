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

# Our custom config template — the official nginx image runs envsubst
# on every file in /etc/nginx/templates/ and writes the result to
# /etc/nginx/conf.d/, so ${API_UPSTREAM} is substituted at container start.
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Copy the Vite build output
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
