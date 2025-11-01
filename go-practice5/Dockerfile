FROM golang:1.25-bookworm AS builder
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOTOOLCHAIN=auto

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 8080
ENV PORT=8080
CMD ["/app/server"]
