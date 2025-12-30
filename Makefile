# Root Makefile - orchestrates subproject Makefiles
.PHONY: help install-all status-all uninstall-all upgrade-all clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Namespaced targets:'
	@echo '  chart-<target>    Run chart management commands (e.g., make chart-install-all)'
	@echo '  api-<target>      Run API commands (e.g., make api-run)'
	@echo ''
	@echo 'Available chart targets:'
	@$(MAKE) -C charts help

# Chart targets - delegate to charts/Makefile
chart-%:
	@$(MAKE) -C charts $*

# API targets - delegate to api/Makefile (future)
api-%:
	@$(MAKE) -C api $*

# Convenience aliases - direct access to common chart commands
install-all: chart-install-all
status-all: chart-status-all
uninstall-all: chart-uninstall-all
upgrade-all: chart-upgrade-all
clean: chart-clean
