FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk --no-cache add bash git make

# deps
COPY go.mod go.sum ./
RUN go mod download

# build
COPY . .
RUN go build -o ./bin/app cmd/main.go

# run
FROM alpine AS runner

COPY --from=builder /app/bin/app /
CMD ["/app"]