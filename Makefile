TAG = $(shell date -u +"%Y%m%d%H%M")

.PHONY: resource-prober

resource-prober:
	docker build --rm -f ./resource-prober.Dockerfile -t registry.cn-shanghai.aliyuncs.com/fusion-app/http-prober:resource-prober.$(TAG) .
	docker push registry.cn-shanghai.aliyuncs.com/fusion-app/http-prober:resource-prober.$(TAG)