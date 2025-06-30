# Variables
MODULE_NAME = github.com/JBrVJxsc/xk6-streamloader
K6_VERSION = v1.0.0
XK6_VERSION = v0.20.1
BUILD_DIR = build
K6_BINARY = $(BUILD_DIR)/k6

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[0;33m
RED = \033[0;31m
NC = \033[0m # No Color

.PHONY: all build clean test test-go test-k6 test-memory help generate-test-files prepare-test-env

# Default target
all: build test

# Help target
help:
	@echo "$(GREEN)Available targets:$(NC)"
	@echo "  $(YELLOW)all$(NC)        - Build extension and run all tests"
	@echo "  $(YELLOW)build$(NC)      - Build k6 binary with streamloader extension"
	@echo "  $(YELLOW)test$(NC)       - Run all tests (Go + k6)"
	@echo "  $(YELLOW)test-go$(NC)    - Run Go unit tests only"
	@echo "  $(YELLOW)test-k6$(NC)    - Run k6 JavaScript tests only"
	@echo "  $(YELLOW)test-memory$(NC) - Run k6 memory comparison tests only"
	@echo "  $(YELLOW)clean$(NC)      - Clean build artifacts"
	@echo "  $(YELLOW)help$(NC)       - Show this help message"

# Build k6 binary with extension
build:
	@echo "$(GREEN)Building k6 with streamloader extension...$(NC)"
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && xk6 build $(K6_VERSION) --with $(MODULE_NAME)=../
	@echo "$(GREEN)✓ Build complete: $(K6_BINARY)$(NC)"

# Setup test environment
prepare-test-env: build
	@echo "$(GREEN)Preparing test environment...$(NC)"
	@mkdir -p $(BUILD_DIR)/testdata/
	@cp testdata/*.csv $(BUILD_DIR)/
	@cp testdata/*.csv $(BUILD_DIR)/testdata/
	@cp testdata/*.json $(BUILD_DIR)/
	@echo "$(GREEN)✓ Test environment prepared$(NC)"

# Run all tests
test: test-go test-k6 test-memory
	@echo "$(GREEN)✓ All tests completed successfully!$(NC)"

# Run Go unit tests
test-go: generate-test-files
	@echo "$(GREEN)Running Go unit tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✓ Go tests completed$(NC)"

# Run k6 JavaScript tests
test-k6: build generate-test-files prepare-test-env
	@echo "$(GREEN)Running k6 JavaScript tests...$(NC)"
	@if [ -f "tests/json/streamloader_k6_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/streamloader_k6_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/streamloader_k6_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/json_utils_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/json_utils_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/json_utils_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/json_roundtrip_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/json_roundtrip_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/json_roundtrip_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/compressed_json_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/compressed_json_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/compressed_json_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/compression_performance_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/compression_performance_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/compression_performance_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/reverse_json_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/reverse_json_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/reverse_json_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/roundtrip_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/roundtrip_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/roundtrip_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/head_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/head_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/head_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/json/tail_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/json/tail_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/json/tail_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/process_csv_test.js" ]; then \
		$(K6_BINARY) run tests/csv/process_csv_test.js; \
	else \
		echo "$(RED)Error: tests/csv/process_csv_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/advanced_process_csv_test.js" ]; then \
		$(K6_BINARY) run tests/csv/advanced_process_csv_test.js; \
	else \
		echo "$(RED)Error: tests/csv/advanced_process_csv_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/edge_case_csv_test.js" ]; then \
		$(K6_BINARY) run tests/csv/edge_case_csv_test.js; \
	else \
		echo "$(RED)Error: tests/csv/edge_case_csv_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/params/test_parameters.js" ]; then \
		$(K6_BINARY) run tests/params/test_parameters.js; \
	else \
		echo "$(RED)Error: tests/params/test_parameters.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/params/parameter_struct_test.js" ]; then \
		$(K6_BINARY) run tests/params/parameter_struct_test.js; \
	else \
		echo "$(RED)Error: tests/params/parameter_struct_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/params/comprehensive_param_test.js" ]; then \
		$(K6_BINARY) run tests/params/comprehensive_param_test.js; \
	else \
		echo "$(RED)Error: tests/params/comprehensive_param_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/params/null_value_param_test.js" ]; then \
		$(K6_BINARY) run tests/params/null_value_param_test.js; \
	else \
		echo "$(RED)Error: tests/params/null_value_param_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/params/struct_tags_test.js" ]; then \
		$(K6_BINARY) run tests/params/struct_tags_test.js; \
	else \
		echo "$(RED)Error: tests/params/struct_tags_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/params/streamloader_fields_test.js" ]; then \
		$(K6_BINARY) run tests/params/streamloader_fields_test.js; \
	else \
		echo "$(RED)Error: tests/params/streamloader_fields_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/orders_products_test.js" ]; then \
		$(K6_BINARY) run tests/csv/orders_products_test.js; \
	else \
		echo "$(RED)Error: tests/csv/orders_products_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/transform_projection_test.js" ]; then \
		$(K6_BINARY) run tests/csv/transform_projection_test.js; \
	else \
		echo "$(RED)Error: tests/csv/transform_projection_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/lazy_quotes_option_test.js" ]; then \
		$(K6_BINARY) run tests/csv/lazy_quotes_option_test.js; \
	else \
		echo "$(RED)Error: tests/csv/lazy_quotes_option_test.js not found$(NC)"; \
		exit 1; \
	fi
	@if [ -f "tests/csv/csv_options_test.js" ]; then \
		$(K6_BINARY) run --quiet tests/csv/csv_options_test.js || exit 1; \
	else \
		echo "$(RED)Error: tests/csv/csv_options_test.js not found$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ k6 tests completed$(NC)"
	@echo "$(YELLOW)Cleaning up temporary test files except our permanent test files...$(NC)"
	@# We're keeping advanced_process.csv and edge_case_test.csv as permanent test files
	@rm -f $(BUILD_DIR)/test_process.csv $(BUILD_DIR)/test_parameters.csv $(BUILD_DIR)/parameter_struct_test.csv $(BUILD_DIR)/comprehensive_param_test.csv $(BUILD_DIR)/null_value_param_test.csv $(BUILD_DIR)/struct_tags_test.csv
	@# Clean up JSON test files
	@rm -f test_output.jsonl test_output.json test_roundtrip.json special_test.json roundtrip_test.jsonl roundtrip_test.json direct_write_test.json large_dataset.json large_dataset_direct.json combined_dataset.json test_reverse_jsonl.json test_reverse_compressed.json multi_roundtrip_test.json
	@# Clean up compressed JSON test files
	@rm -f compressed_output.json direct_compressed.json special_compressed.json medium_compressed.json two_step_result.json direct_result.json

# Run k6 memory test
test-memory: build generate-test-files
	@echo "$(GREEN)Running k6 memory test for built-in open()...$(NC)"
	@# Run k6 in the background, get its PID, and poll its memory usage
	@$(K6_BINARY) run tests/memory/memory_test_open.js > /dev/null 2>&1 & \
	K6_PID=$$!; \
	MAX_RSS=0; \
	while ps -p $$K6_PID > /dev/null; do \
		CURRENT_RSS=$$(ps -p $$K6_PID -o rss= | awk '{print $$1}'); \
		if [ -n "$$CURRENT_RSS" ] && [ $$CURRENT_RSS -gt $$MAX_RSS ]; then \
			MAX_RSS=$$CURRENT_RSS; \
		fi; \
		sleep 0.1; \
	done; \
	wait $$K6_PID; \
	echo "  => Peak memory (RSS) for open(): $$((MAX_RSS / 1024)) MB";

	@echo "$(GREEN)Running k6 memory test for streamloader.loadText()...$(NC)"
	@# Run k6 in the background, get its PID, and poll its memory usage
	@$(K6_BINARY) run tests/memory/memory_test_streamloader.js > /dev/null 2>&1 & \
	K6_PID=$$!; \
	MAX_RSS=0; \
	while ps -p $$K6_PID > /dev/null; do \
		CURRENT_RSS=$$(ps -p $$K6_PID -o rss= | awk '{print $$1}'); \
		if [ -n "$$CURRENT_RSS" ] && [ $$CURRENT_RSS -gt $$MAX_RSS ]; then \
			MAX_RSS=$$CURRENT_RSS; \
		fi; \
		sleep 0.1; \
	done; \
	wait $$K6_PID; \
	echo "  => Peak memory (RSS) for streamloader.loadText(): $$((MAX_RSS / 1024)) MB";
	
	@if [ -f "tests/json/json_memory_test.js" ]; then \
		echo "$(GREEN)Running k6 memory test for JSON utilities...$(NC)"; \
		$(K6_BINARY) run tests/json/json_memory_test.js > /dev/null 2>&1 & \
		K6_PID=$$!; \
		MAX_RSS=0; \
		while ps -p $$K6_PID > /dev/null; do \
			CURRENT_RSS=$$(ps -p $$K6_PID -o rss= | awk '{print $$1}'); \
			if [ -n "$$CURRENT_RSS" ] && [ $$CURRENT_RSS -gt $$MAX_RSS ]; then \
				MAX_RSS=$$CURRENT_RSS; \
			fi; \
			sleep 0.1; \
		done; \
		wait $$K6_PID; \
		echo "  => Peak memory (RSS) for JSON utilities: $$((MAX_RSS / 1024)) MB"; \
	else \
		echo "$(RED)Error: tests/json/json_memory_test.js not found$(NC)"; \
		exit 1; \
	fi;

# Generate large test files
generate-test-files:
	@echo "$(GREEN)Generating large test files...$(NC)"
	@python3 scripts/generate_large_json.py
	@python3 scripts/generate_large_csv.py
	@python3 scripts/generate_large_file.py
	@echo "$(GREEN)✓ Large test files generated$(NC)"

# Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f large.json large.csv large_file.txt
	rm -f test_output.jsonl test_output.json test_roundtrip.json special_test.json
	rm -f roundtrip_test.jsonl roundtrip_test.json direct_write_test.json
	rm -f large_dataset.json large_dataset_direct.json combined_dataset.json
	rm -f compressed_output.json direct_compressed.json special_compressed.json
	rm -f medium_compressed.json two_step_result.json direct_result.json
	@echo "$(GREEN)✓ Clean completed$(NC)"