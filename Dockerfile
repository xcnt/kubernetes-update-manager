FROM golang:1.18-alpine AS build-env

WORKDIR /app

RUN apk update && apk add --no-cache git alpine-sdk bash
RUN apk add --no-cache ca-certificates && update-ca-certificates
RUN adduser -D -g '' appuser
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.1

ADD . /app
RUN make generate_swagger
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o kubernetes-update-manager

FROM alpine:3.15
WORKDIR /app
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /app/kubernetes-update-manager /app/
ENV GIN_MODE release

USER appuser

EXPOSE 9000

ENTRYPOINT [ "/app/kubernetes-update-manager" ]
CMD [ "server" ]
