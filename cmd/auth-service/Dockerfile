FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ./bin/authSvc /app/authSvc

RUN chmod +x /app/authSvc

CMD ["./authSvc"]
