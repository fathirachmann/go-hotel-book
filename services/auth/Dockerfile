FROM golang:1.25-alpine AS build
WORKDIR /workspace

COPY go.work ./
COPY pkg ./pkg
COPY services ./services

WORKDIR /workspace/services/auth
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/auth

FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/service /service
EXPOSE 8001
USER nonroot:nonroot
ENTRYPOINT ["/service"]
