.PHONY: help install-all status-all uninstall-all install-go-ping install-nginx install-prometheus install-grafana upgrade-all

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-all: install-go-ping install-nginx install-prometheus install-grafana ## Install all Helm charts (go-ping, nginx-proxy, prometheus, grafana)
	@echo ""
	@echo "✅ All charts installed!"
	@echo ""
	@$(MAKE) status-all

status-all: ## Check status of all Helm releases
	@echo "📋 Checking status of all releases..."
	@echo ""
	@echo "=== Helm Releases ==="
	@helm list
	@echo ""
	@echo "=== Pod Status ==="
	@kubectl get pods
	@echo ""
	@echo "=== Service Status ==="
	@kubectl get svc
	@echo ""

uninstall-all: uninstall-grafana uninstall-prometheus uninstall-nginx uninstall-go-ping ## Uninstall all Helm charts
	@echo "✅ All charts uninstalled!"

upgrade-all: upgrade-go-ping upgrade-nginx upgrade-prometheus upgrade-grafana ## Upgrade all Helm releases
	@echo "✅ All charts upgraded!"
	@$(MAKE) status-all

install-go-ping: ## Install only go-ping chart
	@echo "🚀 Installing go-ping..."
	@helm install go-ping ./go-ping
	@echo ""

install-nginx: ## Install only nginx-proxy chart
	@echo "🚀 Installing nginx-proxy..."
	@helm install nginx-proxy ./nginx-proxy
	@echo ""

install-prometheus: ## Install only prometheus chart
	@echo "📊 Installing prometheus..."
	@helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
	@helm repo update
	@helm install prometheus prometheus-community/prometheus -f ./prometheus/values.yaml
	@echo ""

install-grafana: ## Install only grafana chart
	@echo "📈 Installing grafana..."
	@helm repo add grafana https://grafana.github.io/helm-charts 2>/dev/null || true
	@helm repo update
	@helm install grafana grafana/grafana -f ./grafana/values.yaml
	@echo ""

upgrade-go-ping: ## Upgrade go-ping chart
	@echo "🔄 Upgrading go-ping..."
	@helm upgrade go-ping ./go-ping

upgrade-nginx: ## Upgrade nginx-proxy chart
	@echo "🔄 Upgrading nginx-proxy..."
	@helm upgrade nginx-proxy ./nginx-proxy

upgrade-prometheus: ## Upgrade prometheus chart
	@echo "🔄 Upgrading prometheus..."
	@helm upgrade prometheus prometheus-community/prometheus -f ./prometheus/values.yaml

upgrade-grafana: ## Upgrade grafana chart
	@echo "🔄 Upgrading grafana..."
	@helm upgrade grafana grafana/grafana -f ./grafana/values.yaml

uninstall-go-ping: ## Uninstall go-ping chart
	@echo "🗑️  Uninstalling go-ping..."
	@helm uninstall go-ping || true

uninstall-nginx: ## Uninstall nginx-proxy chart
	@echo "🗑️  Uninstalling nginx-proxy..."
	@helm uninstall nginx-proxy || true

uninstall-prometheus: ## Uninstall prometheus chart
	@echo "🗑️  Uninstalling prometheus..."
	@helm uninstall prometheus || true

uninstall-grafana: ## Uninstall grafana chart
	@echo "🗑️  Uninstalling grafana..."
	@helm uninstall grafana || true

port-forward: ## Set up port forwarding for all services
	@echo "🔌 Setting up port forwarding..."
	@echo "go-ping:     http://localhost:8080"
	@echo "nginx-proxy: http://localhost:8081"
	@echo "prometheus:  http://localhost:9090"
	@echo "grafana:     http://localhost:3000"
	@echo ""
	kubectl port-forward svc/go-ping 8080:8080 &
	kubectl port-forward svc/nginx-proxy 8081:80 &
	kubectl port-forward svc/prometheus-server 9090:80 &
	kubectl port-forward svc/grafana 3000:80 &
	@echo "✅ Port forwarding active (run 'make stop-port-forward' to stop)"

stop-port-forward: ## Stop all port forwarding
	@echo "⏹️  Stopping port forwarding..."
	@pkill -f "port-forward" || true
	@echo "✅ Port forwarding stopped!"

logs-go-ping: ## Show go-ping logs
	kubectl logs -l app.kubernetes.io/name=go-ping --tail=100 -f

logs-nginx: ## Show nginx-proxy logs
	kubectl logs -l app.kubernetes.io/name=nginx-proxy --tail=100 -f

logs-prometheus: ## Show prometheus logs
	kubectl logs -l app.kubernetes.io/name=prometheus --tail=100 -f

logs-grafana: ## Show grafana logs
	kubectl logs -l app.kubernetes.io/name=grafana --tail=100 -f

clean: uninstall-all ## Uninstall all charts and clean up
	@echo "🧹 Cleaning up..."
	@kubectl delete pods --field-selector status.phase=Failed || true
	@echo "✅ Cleanup complete!"
