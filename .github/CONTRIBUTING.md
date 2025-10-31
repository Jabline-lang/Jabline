## 🤝 **Contributing**

We welcome contributions to make Jabline better:

1. **Fork** the repository
2. **Create** a feature branch
3. **Add** your improvements with tests
4. **Submit** a pull request

### Development Setup
```bash
# Install Go 1.24.6 or later
# Clone and build
go build -o jabline main.go

# Run tests
go test ./...

# Test closures
jabline run test/test.jb
```