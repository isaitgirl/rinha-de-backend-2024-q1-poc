FROM golang:alpine3.19
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN cd /app && \
    ls -la && \
    go build -o /app/main /app/cmd/app.go
CMD ["/app/main"]
