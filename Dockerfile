# Stage 1: Build the frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# Stage 3: Create the final image
FROM alpine:latest
WORKDIR /app

# Copy the built frontend assets from the frontend-builder stage
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Copy the built backend executable from the backend-builder stage
COPY --from=backend-builder /app/main .

# Expose the port the backend listens on
EXPOSE 8888

# Command to run the backend application
# The backend main.go expects DOCKLET_PORT and DOCKLET_HOST_IP (optional)
# We'll rely on the default port 8888 defined in main.go
# If DOCKLET_HOST_IP is needed for some logging or functionality, it might need to be set.
# For now, the backend seems to default it correctly for logging.
CMD ["/app/main"]