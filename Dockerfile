FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git make

WORKDIR /app

ADD . .
RUN make build

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin ./cmd/

FROM scratch

COPY --from=builder ./app/bin .
COPY --from=builder ./app/migrations/ .

EXPOSE 8082

ENTRYPOINT [ "/bin" ]