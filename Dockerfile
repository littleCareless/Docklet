# Stage 1: Install pnpm and setup workspace
FROM node:18-alpine AS workspace-setup

# Install pnpm globally
RUN npm install -g pnpm@8.15.0

WORKDIR /app

# Copy workspace configuration
COPY package.json pnpm-workspace.yaml turbo.json ./
COPY pnpm-lock.yaml* ./

# Copy frontend package.json
COPY frontend/package.json ./frontend/

# Install dependencies for the entire workspace
RUN pnpm install --frozen-lockfile

# Stage 2: Build the frontend
FROM workspace-setup AS frontend-builder

# Copy frontend source code
COPY frontend/ ./frontend/

# Build frontend using turbo
RUN pnpm turbo build --filter=@docklet/frontend

# Stage 3: Build the backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app/backend

# Copy backend dependencies
COPY backend/go.mod backend/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# Copy backend source code
COPY backend/ ./

# Build backend
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# Stage 4: Create the final image
FROM alpine:latest
WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the built frontend assets from the frontend-builder stage
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Copy the built backend executable from the backend-builder stage
COPY --from=backend-builder /app/main .

# Create non-root user for security
RUN addgroup -g 1001 -S docklet && \
    adduser -S docklet -u 1001 -G docklet

# Change ownership of app directory
RUN chown -R docklet:docklet /app

# Switch to non-root user
USER docklet

# Expose the port the backend listens on
EXPOSE 8888

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8888/api/health || exit 1

# Command to run the backend application
CMD ["/app/main"]