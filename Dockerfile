# Build stage
FROM --platform=linux/amd64 golang:1.21.0-alpine AS build

# First add modules list to better utilize caching
COPY go.sum go.mod /src/

WORKDIR /src

COPY . /src


RUN go mod download
RUN go mod tidy

RUN GO111MODULE=on go build -ldflags="-w -s" --tags musl -o my-app main.go

# Run stage
FROM --platform=linux/amd64 ubuntu:20.04
WORKDIR /app

EXPOSE 8080
EXPOSE 9090

COPY --from=build /src/my-app /app/
COPY entrypoint.sh /app/

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["run"]
