FROM golang:1.19-alpine as builder
RUN mkdir -p /build
WORKDIR /build
COPY ./ /build/
WORKDIR /build/cmd/server
ENV CGO_ENABLED=0
RUN go build -o server -ldflags "-s -w"

FROM scratch
EXPOSE 8080
COPY --from=builder /build/cmd/server/server /server
ENTRYPOINT ["/server"]
CMD []