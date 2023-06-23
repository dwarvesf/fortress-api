FROM surnet/alpine-wkhtmltopdf:3.17.0-0.12.6-full as wkhtmltopdf

FROM golang:1.19-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . .

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go install -v ./...
RUN go install -v github.com/rubenv/sql-migrate/sql-migrate@v1.4.0

FROM alpine:3.17.3

RUN echo http://dl-cdn.alpinelinux.org/alpine/edge/community >> /etc/apk/repositories && \
    echo http://dl-cdn.alpinelinux.org/alpine/edge/main >> /etc/apk/repositories && \
    echo http://dl-cdn.alpinelinux.org/alpine/v3.8/main >> /etc/apk/repositories

RUN apk update && \
    apk add --no-cache \
      bash \
      openssl-dev

RUN apk add --no-cache \
  libstdc++ \
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
