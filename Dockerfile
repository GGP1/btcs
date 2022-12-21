FROM golang:1.19-alpine3.16 as builder

WORKDIR /btcs

COPY go.mod .

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go install -ldflags="-s -w"

# ------------------------------------------

FROM alpine:3.16

COPY --from=builder /go/bin/btcs /usr/bin/

COPY --chmod=0755 node/init.sh .

CMD ["./init.sh"]