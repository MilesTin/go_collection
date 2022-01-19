
.Phony: fmt
fmt:
	gofmt -w .
	goimports -w .