build:
	go build -o app cmd/*.go

build-amd64:
	GOOS=linux GOARCH=amd64 go build -o app cmd/*.go

deploy:
	scp app run.sh root@cloud5.gsrco.com.br:/opt/apps/rinha/2/
