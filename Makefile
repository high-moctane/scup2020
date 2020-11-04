TARGET := \
	serialtest \
	serialmotortest \
	serialspeedtest \
	cartpoletest \
	scup

SUBDIR := \
	agent \
	environment \
	logger

RASPI := pi@mocraspizero.local:~/scup2020
RASPI_ENV := GOOS=linux GOARCH=arm GOARM=6

.PHONY: \
	all \
	build \
	build-raspi \
	build-scup-raspi \
	deploy \
	deploy-envs \
	deploy-scup \
	test \
	vet \
	clean

all: build build-raspi

define template_build
	cd bin/$(1) && go build

endef

build:
	$(foreach target,$(TARGET),$(call template_build,$(target)))

define template_build_raspi
	cd bin/$(1) && $(RASPI_ENV) go build -o $(1)-raspi

endef

build-raspi:
	$(foreach target,$(TARGET),$(call template_build_raspi,$(target)))

build-scup-raspi:
	cd bin/scup && $(RASPI_ENV) go build -o scup-raspi

define template_scp
	scp bin/$(1)/$(1)-raspi $(RASPI)/$(target)-raspi

endef

deploy: deploy-envs
	$(foreach target,$(TARGET),$(call template_scp,$(target)))

deploy-envs:
	scp -r envs $(RASPI)

deploy-scup:
	scp bin/scup/scup-raspi  $(RASPI)/scup-raspi

test:
	go test

define template_vet_target
	-cd bin/$(1) && go vet

endef

define template_vet_subdir
	-cd $(1) && go vet

endef

vet:
	-go vet
	$(foreach target,$(TARGET),$(call template_vet_target,$(target)))
	$(foreach subdir,$(SUBDIR),$(call template_vet_subdir,$(subdir)))

clean:
	$(foreach target,$(TARGET),-rm bin/$(target)/$(target))
	$(foreach target,$(TARGET),-rm bin/$(target)/$(target)-raspi)
	$(foreach target,$(TARGET),-rm bin/$(target)/*.result)
