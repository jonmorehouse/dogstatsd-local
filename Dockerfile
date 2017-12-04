FROM golang:latest

COPY . /src

RUN cd /src && CGO_ENABLED=0 GOOS=linux go build -o /dogstatsd-local -a .

FROM scratch

COPY --from=0 /dogstatsd-local .
EXPOSE 8125

ENTRYPOINT ["/dogstatsd-local"]
CMD ["/dogstatsd-local"]
