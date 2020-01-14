FROM golang as builder
WORKDIR /src/

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o bin/app cmd/manager/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /src/bin/app .
ADD config.yaml ./config.yaml
RUN chmod +x app
CMD ["./app"]
