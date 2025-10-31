# Stage 1 — build the binary
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

# Stage 2 — runtime image
FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server /usr/local/bin/server
EXPOSE 8080
ENTRYPOINT ["server"]
