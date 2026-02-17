.PHONY: run-all

run-all:
	@for dir in */ ; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Running example in $$dir"; \
			(cd "$$dir" && go run .) || exit 1; \
		fi \
	done
