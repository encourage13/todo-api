FROM golang:1.27.1 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/api

FROM alpine:3.24
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/server /app/server
COPY config /app/config
CMD ["/app/server"]