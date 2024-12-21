RUNN_CMD := runn run
RUNN_SCENARIO_FILES_MONOLITHIC := $(shell find test/runn/monolithic -name '*.yaml')
RUNN_SCENARIO_FILES_MASTER_SLAVE := $(shell find test/runn/master-slave -name '*.yaml')

api-test-monolithic: $(RUNN_SCENARIO_FILES_MONOLITHIC)
	@if [ -z "$$MONOLITHIC_ENDPOINT" ]; then \
		read -p "Please enter the MONOLITHIC_ENDPOINT: " MONOLITHIC_ENDPOINT; \
		export MONOLITHIC_ENDPOINT; \
		echo "MONOLITHIC_ENDPOINT is set to $$MONOLITHIC_ENDPOINT"; \
	fi; \
	echo "Running runn tests for $$MONOLITHIC_ENDPOINT"; \
	for file in $(RUNN_SCENARIO_FILES_MONOLITHIC); do \
		echo "Running test for $$file"; \
		$(RUNN_CMD) $$file || exit 1; \
	done

api-test-master-slave: $(RUNN_SCENARIO_FILES_MASTER_SLAVE)
	@if [ -z "$$MASTER_ENDPOINT" ] || [ -z "$$SLAVE_ENDPOINT" ]; then \
        read -p "Please enter the MASTER_ENDPOINT: " MASTER_ENDPOINT; \
        read -p "Please enter the SLAVE_ENDPOINT: " SLAVE_ENDPOINT; \
        export MASTER_ENDPOINT SLAVE_ENDPOINT; \
        echo "MASTER_ENDPOINT is set to $$MASTER_ENDPOINT"; \
        echo "SLAVE_ENDPOINT is set to $$SLAVE_ENDPOINT"; \
    fi; \
	echo "Running runn tests for $$MASTER_ENDPOINT and $$SLAVE_ENDPOINT"; \
	for file in $(RUNN_SCENARIO_FILES_MASTER_SLAVE); do \
		echo "Running test for $$file"; \
		$(RUNN_CMD) $$file || exit 1; \
	done

all: test

test: api-test-monolithic api-test-master-slave

.PHONY: all test api-test-monolithic api-test-master-slave