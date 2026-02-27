.PHONY: test

test:
	@echo "Running sqlite example..."
	@go -C sqlite test

	@echo "\nRunning webapi example..."
	@go -C webapi test

	@echo "\nRunning testcontainers mysql example..."
	@go -C testcontainers_mysql test

	@echo "\nRunning testcontainers mongo example..."
	@go -C mongodb_assert test
