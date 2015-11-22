build:
	GO15VENDOREXPERIMENT=1 glide up
	GO15VENDOREXPERIMENT=1 \
	CGO_ENABLED=0 \
	go build -a -tags netgo -installsuffix nocgo -ldflags '-w' .

docker: build
	docker build -t azkube-kvbs .

docker-push: docker
	docker tag -f azkube-kvbs "colemickens/azkube-kvbs:latest"
	docker push "colemickens/azkube-kvbs"

manual-test: quick-build
	./azkube-kvbs \
		-cloudConfigPath=/home/cole/azkube-kvbs_test/azure-config.json \
		-destinationDir=/home/cole/azkube-kvbs_test/output \
		-machineType=master
