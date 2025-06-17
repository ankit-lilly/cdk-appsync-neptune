iLAMBDA_DIRS := $(wildcard lambdas/*)

.PHONY: build

build:
	@for dir in $(LAMBDA_DIRS); do \
		echo "Building $$dir..."; \
		$(MAKE) -C $$dir build; \
	done

