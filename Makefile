default: help

help:
	@echo "RIVA Community Outreach Bot Instructions"
	@echo
	@echo "    - make help             Print help information"
	@echo "    - make build            Build static binary"
	@echo "    - make build-image      Build container image"
	@echo "    - make run-image        Run container image"
	@echo "    - make publish-image    Push container image to docker.io"
	@echo "    - make clean            Clean any built binaries"
	@echo "    - make ssh-prod         Connect to production environment"
	@echo

.PHONY: build
build:
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-static" \
	GOOS=linux \
	GOARCH=amd64 \
	CC="zig cc -target x86_64-linux" \
	CXX="zig c++ -target x86_64-linux" \
	go build -a -o rivabot -ldflags '-extldflags "-static" -w -s' .

.PHONY: build-image
build-image:
	podman build -t docker.io/taronaeo/rivabot -f Dockerfile --format docker .

.PHONY: run-image
run-image:
	podman run -d --name rivabot --restart always -v rivabot:/data docker.io/taronaeo/rivabot

.PHONY: publish-image
publish-image:
	podman push docker.io/taronaeo/rivabot

.PHONY: ssh-prod
ssh-prod:
	@echo
	@echo "    ------------ WARNING WARNING WARNING WARNING WARNING ------------- "
	@echo
	@echo "    You are about to connect to PRODUCTION! Please ensure that this is "
	@echo "    what you intend on doing! Unauthorised access, use, reproduction,  "
	@echo "    possession, modification, interception, damage or transfer         "
	@echo "    (including such attempts) of any content in this system are        "
	@echo "    serious offences under the Computer Misuse Act. If found guilty,   "
	@echo "    an offender can be fined up to SGD100,000 and/or imprisoned up to  "
	@echo "    20 years. If you are not authorised to use this system, DO NOT LOG "
	@echo "    IN OR ATTEMPT TO LOG IN                                            "
	@echo
	@echo "    ------------ WARNING WARNING WARNING WARNING WARNING ------------- "
	@echo
	@echo "    Do you wish to continue? [y/N] " && read ans && [ $${ans:-N } = y ]
	TERM=xterm gcloud compute ssh \
		 		--zone "us-central1-c" \
				"rivalumniops-whatsapp-bot" \
				--project "rivalumniops-whatsapp"

.PHONY: clean
clean:
	rm -f rivabot

