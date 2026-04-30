ARG REGISTRY_URI
FROM ${REGISTRY_URI}/golang:1.24.0 as builder

ARG VERSION=1.0
ARG OS
ARG ARCH



WORKDIR /code
COPY go.mod go.sum /code/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -o ./out/server main.go

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY Makefile ./

FROM ${REGISTRY_URI}/debian:buster-slim

WORKDIR /app

COPY --from=builder /code/out/server ./out/server

COPY config/config.yml config.yml

EXPOSE 8080
ENTRYPOINT ["out/server","-configFile","config.yml"]

# removed some stuff for privacy