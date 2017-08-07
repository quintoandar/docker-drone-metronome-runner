FROM golang:1.7.3
WORKDIR /go/src/github.com/quintoandar/docker-drone-metronome-runner
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:3.6
RUN apk --no-cache add ca-certificates
WORKDIR /usr/local/bin
COPY --from=0 /go/src/github.com/quintoandar/docker-drone-metronome-runner/app .
ENTRYPOINT ["app"]  

