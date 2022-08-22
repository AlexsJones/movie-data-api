VERSION=0.0.1
AWS_BUCKET_NAME ?= "movie-data-api"
AWS_REGION ?= "eu-west-1"

.PHONY: docker-build deploy deploy-fail undeploy postgres up
postgres:
	./scripts/install-postgres.sh
up: postgres deploy
docker-build:
	cd pkg && docker buildx build --build-arg=VERSION=$(VERSION) --platform linux/amd64,linux/arm64,linux/arm . -t cnskunkworks/movie-data-api:$(VERSION) --push
deploy:
	helm upgrade --install movie-data-api . --set=image.tag=$(VERSION) --set=aws.region=$(AWS_REGION) --set=aws.bucket=$(AWS_BUCKET_NAME) --set=postgres.password=$$(kubectl get secrets movie-db-cluster-app -o jsonpath="{.data.password}" | base64 -d)
undeploy:
	helm uninstall movie-data-api
