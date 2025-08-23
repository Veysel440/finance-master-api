FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/base-debian12
ENV ADDR=":8080"
COPY --from=builder /bin/api /api
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/api"]