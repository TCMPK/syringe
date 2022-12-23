# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /syringe

FROM golang:1.19-alpine
ENV GIN_MODE release
WORKDIR /app
COPY --from=build /syringe /syringe
COPY syringe.yml syringe.yml
COPY docs docs
EXPOSE 8000/tcp
ENTRYPOINT [ "/syringe", "syringe.yml" ]