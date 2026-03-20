FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o hitori-backend .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /app/hitori-backend /usr/local/bin/
EXPOSE 8080
CMD ["hitori-backend"]
