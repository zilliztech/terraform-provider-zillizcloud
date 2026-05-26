default:
  @just --list

terralist-build:
  @bash scripts/terralist/build.sh

terralist-validate-parameters:
  @python3 scripts/terralist/validate-publish-parameters.py

terralist-check-version:
  @python3 scripts/terralist/check-version.py \
    --terralist-url "$TERRALIST_URL" \
    --namespace "$TERRALIST_NAMESPACE" \
    --provider-name "$PROVIDER_NAME" \
    --version "$PUBLISH_VERSION"

terralist-setup-gpg-key:
  @bash scripts/terralist/setup-gpg-key-from-secret.sh

terralist-publish:
  @bash scripts/terralist/publish.sh

terralist-publish-all: terralist-build terralist-publish

terralist-verify-install:
  @sh scripts/terralist/verify-terraform-install.sh

terralist-send-feishu-card:
  @python3 scripts/terralist/send-feishu-publish-card.py

terralist-cleanup-gpg-key:
  @bash scripts/terralist/cleanup-gpg-key.sh

_terralist-build-platform:
  @bash scripts/terralist/build-platform.sh

_terralist-check-artifacts:
  @bash scripts/terralist/check-artifacts.sh

_terralist-checksum:
  @bash scripts/terralist/checksum.sh

_terralist-sign:
  @bash scripts/terralist/sign.sh

_terralist-api-key:
  @bash scripts/terralist/api-key.sh

_terralist-ensure-authority:
  @bash scripts/terralist/ensure-authority.sh

_terralist-ensure-key:
  @bash scripts/terralist/ensure-key.sh

_terralist-stage-artifacts:
  @bash scripts/terralist/stage-artifacts.sh

_terralist-upload:
  @bash scripts/terralist/upload.sh

_terralist-verify:
  @bash scripts/terralist/verify.sh

plan dir:
  terraform -chdir={{dir}} plan

refresh dir:
  terraform -chdir={{dir}} refresh

apply dir:
  terraform -chdir={{dir}} apply -auto-approve

destroy dir:
  terraform -chdir={{dir}} destroy -auto-approve


# Run acceptance tests
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

unit-test:
	go test ./... -v 

install:
	go install -v ./...

doc:
	go generate ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

mock-server:
	cd mock-server && go run main.go


prepare: fmt lint doc
