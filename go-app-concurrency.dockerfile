FROM alpine:latest

RUN mkdir /app

COPY goAppConcurrency /app

CMD [ "/app/goAppConcurrency" ]