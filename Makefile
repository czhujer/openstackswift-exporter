IMAGE = registry.mallgroup.com/cc/openstackswift-exporter
VERSION = 1.0

.PHONY: _
_: build publish

.PHONY: build
build:
	docker build -t $(IMAGE):$(VERSION) .

.PHONY: publish
publish:
	docker push $(IMAGE):$(VERSION)
