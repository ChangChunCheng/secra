# Stage 0: Generate TypeScript types from Protobuf using buf image
FROM bufbuild/buf:latest AS proto-gen
WORKDIR /workspace
COPY api ./api
WORKDIR /workspace/api
RUN buf generate --template buf.gen.yaml

# Stage 1: Build the Next.js application
FROM node:20-alpine AS builder
WORKDIR /app

# Copy package files and install dependencies
COPY frontend/package*.json ./
RUN npm install

# Copy frontend source code
COPY frontend .

# Copy generated protobuf types from proto-gen stage
COPY --from=proto-gen /workspace/frontend/src/lib/gen ./src/lib/gen

# Set build-time environment variable
ARG NEXT_PUBLIC_API_URL=/api/v1
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Build the Next.js app (without prebuild script since we already generated types)
RUN npm run build

# Stage 2: Serve static files AND proxy API via Nginx
FROM nginx:alpine
COPY --from=builder /app/out /usr/share/nginx/html

# Unified Nginx Configuration
RUN echo 'server { \
    listen 80; \
    server_name localhost; \
    \
    location / { \
        root /usr/share/nginx/html; \
        try_files $uri $uri.html /index.html; \
    } \
    \
    # API Proxy: Explicitly forward to backend on 8080 \
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
