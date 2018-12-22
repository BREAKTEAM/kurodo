# Stage: Dev
FROM golang:alpine AS builder
WORKDIR /go/src/github.com/BREAKTEAM/kurodo/

COPY . .

RUN apk update && apk add git
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -o kurodo

# Stage: Pro
FROM alpine:latest

COPY --from=builder /go/src/github.com/BREAKTEAM/kurodo/kurodo /usr/local/bin/
WORKDIR /usr/local/bin

ENTRYPOINT [ "kurodo" ]
CMD [ "" ]
