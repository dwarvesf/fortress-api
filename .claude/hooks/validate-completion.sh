#!/bin/bash

echo 'Tests below need to pass, DONT skip them when they fail!' >&2

# Create output directory if it doesn't exist
mkdir -p .claude/validation-outputs

# Run tests
echo "Running tests..." >&2
make test > .claude/validation-outputs/test-output.log 2>&1
if [ $? -ne 0 ]; then
echo "❌ Tests failed. Check .claude/validation-outputs/test-output.log for details." >&2
echo "Failed test lines:" >&2
grep -i "failed\|error\|✗\|❌\|FAIL" .claude/validation-outputs/test-output.log | head -10 >&2
exit 2  # Exit code 2 blocks and triggers retry
else
echo "✅ Tests passed" >&2
fi

# Run linting
echo "Running linting..." >&2
make lint > .claude/validation-outputs/lint-output.log 2>&1
if [ $? -ne 0 ]; then
echo "❌ Linting failed. Check .claude/validation-outputs/lint-output.log for details." >&2
echo "Failed lint lines:" >&2
grep -i "failed\|error\|✗\|❌" .claude/validation-outputs/lint-output.log | head -10 >&2
exit 2
else
echo "✅ Linting passed" >&2
fi

# Run build to check for compilation errors
echo "Running build..." >&2
make build > .claude/validation-outputs/build-output.log 2>&1
if [ $? -ne 0 ]; then
echo "❌ Build failed. Check .claude/validation-outputs/build-output.log for details." >&2
echo "Failed build lines:" >&2
grep -i "failed\|error\|✗\|❌" .claude/validation-outputs/build-output.log | head -10 >&2
exit 2
else
echo "✅ Build passed" >&2
fi

echo "🎉 All checks passed!" >&2
exit 0  # Success - allows Claude to proceed