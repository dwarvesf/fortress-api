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
    xfonts-base \
    libjpeg62-turbo \
    tzdata \
    wget && \
  rm -rf /var/lib/apt/lists/*

# Install wkhtmltopdf with patched Qt (required for selectable text in PDFs)
# DO NOT use apt-get install wkhtmltopdf - it installs unpatched version that rasterizes text
RUN wget -q https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.bookworm_amd64.deb && \
  dpkg -i wkhtmltox_0.12.6.1-3.bookworm_amd64.deb && \
  rm wkhtmltox_0.12.6.1-3.bookworm_amd64.deb

RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime

WORKDIR /
COPY --from=builder /go/bin/* /usr/bin/
COPY migrations /migrations
COPY dbconfig.yml /
COPY pkg/templates /templates

ENTRYPOINT [ "server" ]
