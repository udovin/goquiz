FROM golang:1.18-alpine AS build
RUN apk add --no-cache git gcc linux-headers libc-dev
WORKDIR /src/goquiz
COPY go.mod go.sum /src/goquiz/
RUN go mod download -x
COPY . /src/goquiz
RUN go build -o goquiz .

FROM alpine
RUN apk add --no-cache curl
RUN addgroup -S goquiz -g 1000 && adduser -S goquiz -G goquiz -u 1000
COPY --from=build /src/goquiz/goquiz /bin/goquiz
USER goquiz
ENTRYPOINT ["/bin/goquiz"]
