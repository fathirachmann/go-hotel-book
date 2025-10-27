FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/catalog


FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/service /service
EXPOSE 8002
USER nonroot:nonroot
ENTRYPOINT ["/service"]