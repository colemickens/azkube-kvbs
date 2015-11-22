build:
	GO15VENDOREXPERIMENT=1 glide up
	GO15VENDOREXPERIMENT=1 \
	CGO_ENABLED=0 \
	go build -a -tags netgo -installsuffix nocgo -ldflags '-w' .

docker: build
	docker build -t azkvbs .

docker-push: docker
	docker tag -f azkvbs "colemickens/azkvbs:latest"
	docker push "colemickens/azkvbs"

quick-build:
	
	go build .

manual-test: quick-build
	./azkvbs \
		-cloudConfigPath=/home/cole/azkvbs_test/azure-config.json \
		-destinationDir=/home/cole/azkvbs_test/output \
		-machineType=master
