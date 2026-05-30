FROM golang:1.20-alpine AS build
RUN apk add --no-cache git build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /app/bin/server /usr/local/bin/server
WORKDIR /app
ENV PORT=8080
EXPOSE 8080
CMD ["/usr/local/bin/server"]
