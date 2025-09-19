FROM surnet/alpine-wkhtmltopdf:3.17.0-0.12.6-full AS wkhtmltopdf

FROM golang:1.25 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build  -v ./cmd/server ./cmd/mcp-server
RUN go install -v ./cmd/server ./cmd/mcp-server
RUN go install -v github.com/rubenv/sql-migrate/sql-migrate@v1.4.0

FROM alpine:3.18 AS runtime

RUN echo http://dl-cdn.alpinelinux.org/alpine/edge/community >> /etc/apk/repositories && \
  echo http://dl-cdn.alpinelinux.org/alpine/edge/main >> /etc/apk/repositories && \
  echo http://dl-cdn.alpinelinux.org/alpine/v3.8/main >> /etc/apk/repositories

RUN apk update && \
  apk add --no-cache \
  bash \
  openssl-dev \
  tzdata

RUN apk add --no-cache \
  libc6-compat \
  libstdc++ \
  libgcc \
  libx11 \
  libxrender \
  libxext \
  libssl1.1 \
  ca-certificates \
  fontconfig \
  freetype \
  ttf-dejavu \
  ttf-droid \
  ttf-freefont \
  ttf-liberation \
  ttf-ubuntu-font-family

RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime

COPY --from=wkhtmltopdf /bin/wkhtmltopdf /usr/bin/wkhtmltopdf
RUN chmod +x /usr/bin/wkhtmltopdf

WORKDIR /
COPY --from=builder /go/bin/* /usr/bin/
COPY migrations /migrations
COPY dbconfig.yml /
COPY pkg/templates /templates

ENTRYPOINT [ "server" ]
