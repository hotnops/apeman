FROM golang:1.20-alpine AS builder
WORKDIR /src
COPY go.mod go.sum go.work ./
COPY . .
RUN go mod download
RUN go build -ldflags '-s -w' -o /app ./api/main.go
# Second stage: copy the binary into a clean image.
FROM scratch
COPY --from=builder /app /app
COPY dawgsConfig.json /dawgsConfig.json
ENTRYPOINT ["/app"]