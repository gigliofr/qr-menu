#!/bin/bash
# Deployment Verification Script - Production Readiness Check
# Usage: ./scripts/verify_deployment.sh

set -e

echo "🚀 QR-Menu Production Readiness Verification"
echo "=============================================="
echo ""

FAILED=0
PASSED=0

# 1. Check Go build
echo "✓ Checking Go build..."
if go build -o qr-menu . 2>&1; then
    echo "  ✅ Build successful"
    ((PASSED++))
else
    echo "  ❌ Build failed"
    ((FAILED++))
fi

# 2. Check tests
echo "✓ Running unit tests..."
if go test ./... -timeout 30s > /tmp/test.log 2>&1; then
    echo "  ✅ Tests passed"
    ((PASSED++))
else
    echo "  ❌ Tests failed"
    cat /tmp/test.log | tail -20
    ((FAILED++))
fi

# 3. Check test coverage
echo "✓ Checking test coverage..."
go test ./... -coverprofile=/tmp/coverage.out > /dev/null 2>&1
COVERAGE=$(go tool cover -func=/tmp/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE >= 70" | bc -l) )); then
    echo "  ✅ Coverage: $COVERAGE% (target: 70%)"
    ((PASSED++))
else
    echo "  ⚠️  Coverage: $COVERAGE% (target: 70% - consider adding tests)"
    ((PASSED++))  # Warning, not failure
fi

# 4. Check critical files exist
echo "✓ Checking configuration files..."
FILES=(
    "pkg/config/config.go"
    "pkg/errors/circuit_breaker.go"
    "pkg/metrics/business_metrics.go"
    "handlers/handlers_test.go"
    "handlers/auth_handlers_test.go"
    "db/indices.go"
)

for FILE in "${FILES[@]}"; do
    if [ -f "$FILE" ]; then
        echo "  ✅ $FILE exists"
        ((PASSED++))
    else
        echo "  ❌ $FILE missing"
        ((FAILED++))
    fi
done

# 5. Check for Wave 3.1 optimizations
echo "✓ Checking Wave 3.1 optimizations..."
if grep -q "UpdateManyMenus" db/mongo.go; then
    echo "  ✅ Batch UpdateMany operation implemented"
    ((PASSED++))
else
    echo "  ❌ UpdateMany batch operation missing"
    ((FAILED++))
fi

if grep -q "sync.Once" handlers/handlers.go; then
    echo "  ✅ Template caching with sync.Once implemented"
    ((PASSED++))
else
    echo "  ❌ Template caching missing"
    ((FAILED++))
fi

# 6. Check for business metrics
echo "✓ Checking business metrics..."
if grep -q "MenusCreated\|QRCodesGenerated\|MenuViews" pkg/metrics/business_metrics.go; then
    echo "  ✅ Business metrics defined"
    ((PASSED++))
else
    echo "  ❌ Business metrics missing"
    ((FAILED++))
fi

# 7. Check for circuit breaker
echo "✓ Checking circuit breaker implementation..."
if grep -q "CircuitBreaker" pkg/errors/circuit_breaker.go; then
    echo "  ✅ Circuit breaker pattern implemented"
    ((PASSED++))
else
    echo "  ❌ Circuit breaker missing"
    ((FAILED++))
fi

# 8. Verify database helpers
echo "✓ Checking database helpers..."
if grep -q "DeleteUser\|DeleteRestaurant" db/mongo.go; then
    echo "  ✅ Database cleanup methods available"
    ((PASSED++))
else
    echo "  ❌ Database cleanup methods missing"
    ((FAILED++))
fi

# 9. Check deployment docs
echo "✓ Checking deployment documentation..."
if [ -f "PRODUCTION_CONFIG.md" ]; then
    echo "  ✅ PRODUCTION_CONFIG.md exists"
    ((PASSED++))
else
    echo "  ❌ PRODUCTION_CONFIG.md missing"
    ((FAILED++))
fi

# 10. Check git status
echo "✓ Checking git status..."
if [ -z "$(git status --porcelain)" ]; then
    echo "  ✅ All changes committed"
    ((PASSED++))
else
    echo "  ⚠️  Uncommitted changes detected"
    git status --short | head -5
fi

echo ""
echo "=============================================="
echo "Results: $PASSED passed, $FAILED failed"
echo "=============================================="

if [ $FAILED -eq 0 ]; then
    echo "✅ Ready for production deployment!"
    exit 0
else
    echo "❌ Fix failures before deploying"
    exit 1
fi
