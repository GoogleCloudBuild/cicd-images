GOLANGCI_VERSION = v1.56.2
BIN = $(CURDIR)/.bin
GO = go
Q = $(if $(filter 1,$V),,@)


.PHONY: golangci-lint
golangci-lint: ; $(info Running golangci-lint…) @ ## Run golangci-lint
	$Q golangci-lint run

.PHONY: fmt
fmt: ; $(info $(M) Running gofmt…) @ ## Run gofmt
	$Q $(GO) fmt  ./...


.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(BIN)
	@rm -rf bin

