FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .

EXPOSE 8080

RUN apt-get update && apt-get install -y git

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$(git describe --tags --always)" -o /arevbond ./cmd/arevbond/main.go

FROM alpine AS build-release-stage

RUN apk add --no-cache tzdata
ENV TZ="Europe/Moscow"

WORKDIR /

COPY --from=build-stage /arevbond /arevbond

EXPOSE 8080

ENTRYPOINT ["/arevbond"]