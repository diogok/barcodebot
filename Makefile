all: build

clean: 
	rm barcodebot-*

barcodebot-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -tags netgo -ldflags '-w' -o barcodebot-arm

barcodebot-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o barcodebot-amd64

amd64: barcodebot-amd64

arm: barcodebot-arm

docker: amd64
	cp barcodebot-amd64 barcodebot
	docker build -t diogok/barcodebot .
	rm barcodebot

docker-arm: arm
	cp barcodebot-arm barcodebot
	docker build -t diogok/barcodebot:arm .
	rm barcodebot

push:
	docker push diogok/barcodebot

