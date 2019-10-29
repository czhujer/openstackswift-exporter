IMAGE = registry.mallgroup.com/cc/openstackswift-exporter
VERSION = 1.0

.PHONY: _
_: build publish

.PHONY: build
build:
	#@test -n "$(GITLAB_TOKEN)" || (echo "Set GITLAB_TOKEN" && exit 1)
	#cp config.go config.go.bak
	#echo "package main\n\nconst gitlabToken = \"$(GITLAB_TOKEN)\"" > config.go
	docker build -t $(IMAGE):$(VERSION) .
	#mv config.go.bak config.go

.PHONY: publish
publish:
	docker push $(IMAGE):$(VERSION)
