FROM golang:1.12 as build
ARG binary
WORKDIR /go/src/github.com/nagypeterjob-edu/service-telemetry
ADD . /go/src/github.com/nagypeterjob-edu/service-telemetry
RUN cd /go/src/github.com/nagypeterjob-edu/service-telemetry && make build-${binary}

FROM gcr.io/distroless/base
COPY --from=build /go/src/github.com/nagypeterjob-edu/service-telemetry/bin/service /app
CMD ["/app"]