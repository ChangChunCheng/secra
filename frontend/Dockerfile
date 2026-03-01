# Stage 1: Build the React application
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
# Build and export to /app/out
RUN npm run build

# Stage 2: Serve static files AND proxy API via Nginx
FROM nginx:alpine
COPY --from=builder /app/out /usr/share/nginx/html

# Unified Nginx Configuration: Static Files + API Proxy
RUN echo 'server { \
    listen 80; \
    server_name localhost; \
    \
    # Frontend - Static Files \
    location / { \
        root /usr/share/nginx/html; \
        try_files $uri $uri.html /index.html; \
    } \
    \
    # Backend API Proxy - Forward to Go Server \
    location /api/ { \
        proxy_pass http://server:8080; \
        proxy_set_header Host $host; \
        proxy_set_header X-Real-IP $remote_addr; \
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; \
        proxy_set_header X-Forwarded-Proto $scheme; \
    } \
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
