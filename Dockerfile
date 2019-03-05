FROM golang:alpine AS build-env

WORKDIR /app

ENV GOFLAGS -mod=vendor 
RUN apk add --update git alpine-sdk bash
ADD . /app
RUN go build -o kubernetes-update-manager

FROM alpine:3.9
WORKDIR /app
COPY --from=build-env /app/kubernetes-update-manager /app/

CMD [ "kubernetes-update-manager" ]
