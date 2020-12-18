FROM golang:alpine AS build

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /service

COPY ./go.mod ./go.sum ./
RUN GOPROXY=https://proxy.golang.org go mod download

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags="-w -s" \
	-installsuffix "static" \
	-o /build/service /service/cmd/service

FROM alpine:latest

RUN apk update && apk add ca-certificates

COPY --from=build /build/service /service

EXPOSE 8080

ENTRYPOINT ["/service"]