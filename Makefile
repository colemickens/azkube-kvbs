build:
	CGO_ENABLED=0 go build -a -installsuffix nocgo .

docker: build
	docker build -t azkvbs .

docker-push: docker
	docker tag -f azkvbs "colemickens/azkvbs:latest"
	docker push "colemickens/azkvbs"

manual-test:
	docker run \
		-v "`pwd`/testdata/waagent:/var/lib/waagent" \
		-v "/etc/ssl/certs:/etc/ssl/certs" \
		-v "`pwd`/testdata/kubernetes:/etc/kubernetes" \
		"colemickens/azkvbs" "/azkvbs"
