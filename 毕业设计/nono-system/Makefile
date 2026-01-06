.PHONY: help build test clean oracle backend frontend deploy

help:
	@echo "可用命令:"
	@echo "  make oracle    - 构建并运行预言机服务"
	@echo "  make backend   - 构建并运行后端API服务"
	@echo "  make frontend  - 构建并运行前端应用"
	@echo "  make build     - 构建所有服务"
	@echo "  make test      - 运行所有测试"
	@echo "  make clean     - 清理构建产物"

build:
	@echo "构建所有服务..."
	cd oracle && go build -o ../bin/oracle ./cmd/oracle
	cd backend && go build -o ../bin/backend ./cmd/server

oracle:
	@echo "启动预言机服务..."
	cd oracle && go run ./cmd/oracle

backend:
	@echo "启动后端API服务..."
	cd backend && go run ./cmd/server

test:
	@echo "运行测试..."
	go test ./...

clean:
	@echo "清理构建产物..."
	rm -rf bin/
	rm -rf dist/
	find . -name "*.test" -delete

