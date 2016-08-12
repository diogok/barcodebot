FROM alpine:3.1

RUN apk add --update ca-certificates
COPY ./barcodebot /opt/barcodebot
CMD ["/opt/barcodebot"]


