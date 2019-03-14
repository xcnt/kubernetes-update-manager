FROM golang:alpine AS build-env

WORKDIR /app

ENV GOFLAGS -mod=vendor 
RUN apk add --update git alpine-sdk bash
RUN go get github.com/swaggo/swag/cmd/swag

ADD . /app
RUN swag init --generalInfo web/router.go --dir ./web --swagger web/docs/swagger/
RUN go build -o kubernetes-update-manager

FROM alpine:3.9
WORKDIR /app
COPY --from=build-env /app/kubernetes-update-manager /app/

CMD [ "kubernetes-update-manager" ]
