# syntax=docker/dockerfile:1

FROM golang:alpine as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /syringe
RUN mkdir /etc/syringe && touch /etc/syringe/domains

FROM scratch
WORKDIR /app
COPY --from=build /syringe /syringe
COPY --from=build /etc/syringe /etc/syringe
COPY syringe.yml syringe.yml
COPY docs docs
EXPOSE 8000/tcp
ENTRYPOINT [ "/syringe", "syringe.yml" ]