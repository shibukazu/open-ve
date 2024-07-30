CONFIG_FILE := config.yaml
HTTP_ADDR := "0.0.0.0:8080"
# Include HTTP_ADDR from config.yaml
ifneq (,$(wildcard $(CONFIG_FILE)))
    ifneq (,$(shell test -s $(CONFIG_FILE)))
        ifeq (0,$(shell yq e . $(CONFIG_FILE) > /dev/null 2>&1;))
            TMP_HTTP_ADDR := $(shell yq e .http.addr $(CONFIG_FILE))
            ifneq ($(TMP_HTTP_ADDR),null)
                HTTP_ADDR := $(TMP_HTTP_ADDR)
            endif
        endif
    endif
endif

ENDPOINT := "http://$(subst ",,$(HTTP_ADDR))"
$(eval export ENDPOINT)

RUNN_CMD := runn run
RUNN_SCENARIO_FILES := $(shell find runn -name '*.yaml')

all: test

test: $(RUNN_SCENARIO_FILES)
	@echo "Running runn tests for $(ENDPOINT)"
	@for file in $(RUNN_SCENARIO_FILES); do \
		echo "Running test for $$file"; \
		$(RUNN_CMD) $$file || exit 1; \
	done

.PHONY: all test