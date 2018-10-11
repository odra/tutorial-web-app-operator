REG=docker.io
ORG=integreatly
IMAGE=tutorial-web-app-operator
TAG=latest

go-build:
	go build -o  tmp/_output/bin/tutorial-web-app-operator cmd/tutorial-web-app-operator/main.go

template-copy:
	mkdir -p tmp/_output/deploy/template
	cp deploy/template/tutorial-web-app.yml tmp/_output/deploy/template

docker-build:
	docker build . -f tmp/build/Dockerfile -t ${REG}/${ORG}/${IMAGE}:${TAG}

build: go-build template-copy docker-build

docker-push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}
