IMAGE_NAME ?= "tidb-hack/movie-catalogue:v1"
.PHONY: run
run:
	go build -o bin/movie-catalogue-api
	./bin/movie-catalogue-api
docker-start:
	docker build -t $(IMAGE_NAME) .
	docker run -p 8080:8081 $(IMAGE_NAME)
docker-push:
	docker push $(IMAGE_NAME)
docker-run:
	docker run -p 8081:8081 $(IMAGE_NAME)
deploy:
	cd chart && helm upgrade --install movie-catalogue . --set=image.tag="v3" --set=postgres.password=$$(kubectl get secrets movie-db-cluster-app -o jsonpath="{.data.password}" | base64 --decode) && cd ../
