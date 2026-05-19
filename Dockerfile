FROM node:20-alpine AS frontend-builder
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o litewiki .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates git
WORKDIR /app
COPY --from=backend-builder /build/backend/litewiki .
# 前端先放到镜像的 /app/frontend-dist，启动脚本会复制到 /srv/admin
COPY --from=frontend-builder /build/frontend/dist /app/frontend-dist
COPY deploy-entrypoint.sh /app/deploy-entrypoint.sh
RUN chmod +x /app/deploy-entrypoint.sh
EXPOSE 8080
CMD ["/app/deploy-entrypoint.sh"]
