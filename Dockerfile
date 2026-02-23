FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o gtd-cli ./cmd/gtd-cli

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/gtd-cli /usr/local/bin/gtd-cli

RUN chmod +x /usr/local/bin/gtd-cli

ENTRYPOINT ["gtd-cli"]
CMD ["--help"]
