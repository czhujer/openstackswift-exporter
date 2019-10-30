IMAGE = registry.mallgroup.com/cc/openstackswift-exporter
VERSION = 1.1

.PHONY: _
_: build publish

.PHONY: build
build:
	docker build -t $(IMAGE):$(VERSION) .

.PHONY: publish
publish:
	docker push $(IMAGE):$(VERSION)
