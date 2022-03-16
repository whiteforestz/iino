build:
	docker build . -t iino:1.0

run: build
	docker run --name iino -d -v /proc:/tmp/inno/proc iino:1.0

remove: stop
	docker rm iino

stop:
	docker stop iino
