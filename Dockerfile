FROM golang:bullseye AS builder

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -mod readonly -v -o sogorro

FROM debian:buster-slim
RUN set -eux && \
    apt-get update && \
    apt-get install -qy ca-certificates && \
    rm -fr /var/lib/apt/lists/*

COPY --from=builder /app/sogorro /app/sogorro

EXPOSE 8080

CMD ["/app/sogorro"]