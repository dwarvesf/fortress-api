FROM golang:1.25-bookworm AS builder


WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go install -v ./...
RUN CGO_ENABLED=0 go install -v github.com/rubenv/sql-migrate/sql-migrate@v1.4.0

FROM debian:bookworm-slim AS runtime

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    fontconfig \
    fonts-dejavu-core \
    fonts-droid-fallback \
    fonts-freefont-ttf \
    fonts-liberation \
    fonts-noto-core \
    libfreetype6 \
    libx11-6 \
    libxext6 \
    libxrender1 \
    libssl3 \
    libstdc++6 \
    xfonts-75dpi \
    tzdata \
    wkhtmltopdf && \
  rm -rf /var/lib/apt/lists/*

RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime

WORKDIR /
COPY --from=builder /go/bin/* /usr/bin/
COPY migrations /migrations
COPY dbconfig.yml /
COPY pkg/templates /templates

ENTRYPOINT [ "server" ]
