NAME=iap-gateway
VERSION=1.0.0
PORT=8091
REGISTRY_PREFIX=$(if $(REGISTRY),$(addsuffix /, $(REGISTRY)))

.PHONY: build update rollback create publish

build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags "-X main.Version=${VERSION}"
	docker build -t ${NAME}:${VERSION} .

publish:
	docker tag ${NAME}:${VERSION} ${REGISTRY_PREFIX}${NAME}:${VERSION}
	docker push ${REGISTRY_PREFIX}${NAME}:${VERSION}

update:
	docker service update --image ${NAME}:${VERSION} ${NAME}

rollback:
	docker service rollback ${NAME}

create:
	docker service create --replicas 1 -p ${PORT}:${PORT} \
		--env "HOST={{.Node.Hostname}}" \
		--name ${NAME} ${REGISTRY_PREFIX}${NAME}:${VERSION}
