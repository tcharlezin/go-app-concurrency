FROM alpine:latest

RUN mkdir /app
RUN mkdir /templates

COPY myapp /app
COPY cmd/web/templates /templates

CMD [ "/app/myapp" ]