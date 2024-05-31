
FROM --platform=linux/amd64 golang:1.21.0-alpine AS build

WORKDIR /src

COPY . /src

# RUN go mod download because source code is updated

RUN go build -ldflags="-w -s" -o my-app main.go

FROM --platform=linux/amd64 ubuntu:20.04
WORKDIR /app

EXPOSE 8080
EXPOSE 9090

COPY --from=build /src/my-app /app/

CMD ["./my-app", "run"]
