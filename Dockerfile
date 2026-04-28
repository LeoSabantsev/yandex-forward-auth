# This is a multi-stage Dockerfile and requires >= Docker 17.05
FROM golang:1.26.2-bookworm AS builder

ENV GOPROXY=https://proxy.golang.org,direct
ENV GO111MODULE=on

WORKDIR /src/app

# Install the same Buffalo CLI version used locally.
RUN go install github.com/gobuffalo/cli/cmd/buffalo@v0.18.14

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN buffalo build --static -o /bin/app

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /bin

COPY --from=builder /bin/app /bin/app

ENV GO_ENV=production
ENV ADDR=0.0.0.0
ENV PORT=3000

EXPOSE 3000

CMD ["/bin/app"]