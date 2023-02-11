FROM golang:1.19 as build

#ENV GOPROXY=https://goproxy.cn
#ENV GO111MODULE=on

WORKDIR /go/cache

ADD go.mod .
ADD go.sum .
RUN go mod download

WORKDIR /go/release


ADD . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o wxAutoSave main.go


FROM ubuntu
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/release/wxAutoSave /bin/wxAutoSave


ENTRYPOINT ["/bin/wxAutoSave"]
