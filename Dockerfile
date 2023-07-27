## Build stage
FROM golang:1.17-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o movie-catalogue .

## Deploy stage
FROM alpine

COPY --from=build /src/movie-catalogue /usr/local/bin/movie-catalogue

WORKDIR /usr/local/bin

EXPOSE 8080

CMD [ "./movie-catalogue" ]

