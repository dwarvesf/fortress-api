FROM golang:1.19-alpine
RUN mkdir /build
WORKDIR /build
COPY . .

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go mod download
RUN go install -v github.com/rubenv/sql-migrate/sql-migrate@v1.2.0
RUN go install -v github.com/golang/mock/mockgen@v1.6.0
RUN go install -v github.com/vektra/mockery/v2@v2.15.0
RUN go install -v github.com/swaggo/swag/cmd/swag@v1.8.7
RUN go install -v github.com/cosmtrek/air@latest

FROM golang:1.19-alpine
RUN apk --no-cache add ca-certificates
RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0

WORKDIR /
COPY --from=0 /go/bin/* /usr/bin/
COPY --from=0 /go/pkg/ /go/pkg/
