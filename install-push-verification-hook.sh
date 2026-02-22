#!/bin/bash
# install-push-verification-hook.sh
# Installs the pre-push verification hook in a git repository
# Usage: ./install-push-verification-hook.sh [repository-path]
#
# This hook requires human confirmation before any git push proceeds.
# It prevents AI agents from pushing code without explicit human approval.

set -e

REPO_DIR="${1:-.}"

# Check if we're in a git repository
if [ ! -d "$REPO_DIR/.git" ]; then
    echo "Error: Not a git repository: $REPO_DIR"
    echo "Usage: $0 [repository-path]"
    exit 1
fi

HOOK_DIR="$REPO_DIR/.git/hooks"
HOOK_FILE="$HOOK_DIR/pre-push"

# Create the hook
cat > "$HOOK_FILE" << 'HOOK_SCRIPT'
#!/bin/bash
# Pre-push verification hook
# Requires human confirmation before any push proceeds
# This prevents AI agents from pushing without explicit human approval

# Generate random 6-character code
CODE="PUSH-$(head /dev/urandom | LC_ALL=C tr -dc 'A-Z0-9' | head -c 6)"

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  PUSH VERIFICATION REQUIRED"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  Enter this code to confirm push: $CODE"
echo ""
read -p "  Verification code: " INPUT

if [ "$INPUT" != "$CODE" ]; then
    echo ""
    echo "  ✗ Code mismatch. Push cancelled."
    echo ""
    exit 1
fi

echo ""
echo "  ✓ Verified. Proceeding with push."
echo ""
exit 0
HOOK_SCRIPT

# Make executable
chmod +x "$HOOK_FILE"

echo "✓ Push verification hook installed in: $REPO_DIR"
echo ""
echo "To test: git push origin main --dry-run"
echo "To remove: rm $HOOK_FILE"
