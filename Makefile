build:
	docker build -t wukongchatserver .
push:
	docker tag wukongchatserver wukongim/wukongchatserver:latest
	docker push wukongim/wukongchatserver:latest
deploy:
	docker build -t wukongchatserver .
	docker tag wukongchatserver wukongim/wukongchatserver:latest
	docker push wukongim/wukongchatserver:latest
run-dev:
	docker-compose build;docker-compose up -d
stop-dev:
	docker-compose stop
env-test:
	docker-compose -f ./testenv/docker-compose.yaml up -d 