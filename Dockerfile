FROM golang:1.19 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./main.go ./
COPY ./api ./api
RUN CGO_ENABLED=0 GOOS=linux go build -o /uni-passau-bot

# Run the tests in the container
FROM build AS run-test
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build /uni-passau-bot /uni-passau-bot

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/uni-passau-bot"]
