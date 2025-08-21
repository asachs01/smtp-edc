# Project Cleanup Recommendations

## Completed Cleanup Actions

1. **README.md Updates**
   - ✅ Removed duplicate Homebrew tap documentation (lines 200-279)
   - ✅ Fixed incorrect GitHub username in Go install path (asachs → asachs01)
   - ✅ Added missing `-tags cli` flag to build command
   - ✅ Reorganized documentation section for clarity

2. **Gitignore Updates**
   - ✅ Added compiled binaries (smtp-edc, smtp-edc-ui)
   - ✅ Added bin/ directory
   - ✅ Added empty pkg directories
   - ✅ Added Python virtual environment (.venv/)
   - ✅ Added backup file patterns (*~)

3. **Created Cleanup Script**
   - ✅ Added `scripts/cleanup.sh` for regular maintenance

## Files/Directories to Remove

Run the cleanup script to remove these items:
```bash
./scripts/cleanup.sh
```

This will remove:
- `integration_test.go.bak` - Old backup file
- `test_suite_comprehensive.go.bak` - Old backup file
- `smtp-edc` - Compiled binary (shouldn't be in git)
- `smtp-edc-ui` - Compiled binary (shouldn't be in git)
- Empty directories in `pkg/`
- Frontend build caches

## Documentation Inconsistencies

1. **docs/MCP_TOOLS_REFERENCE.md**
   - This appears to be old conceptual documentation for MCP
   - Conflicts with the new MCP implementation in PR #3
   - Recommendation: Archive or remove after MCP PR is merged

2. **Missing Documentation**
   - No docs/SMTP_SECURITY_GUIDE.md (referenced in old README)
   - Consider creating security documentation

## Additional Recommendations

1. **Directory Structure**
   - The `pkg/` directory with empty subdirectories should be removed if not used
   - Consider removing `.venv/` Python virtual environment if not needed

2. **Build Artifacts**
   - Ensure all build outputs are properly gitignored
   - Regular cleanup of build/ and bin/ directories

3. **Workflow Files**
   - Review `.github/workflows/` for any outdated CI/CD configurations
   - Ensure release workflow handles new MCP components

4. **Dependencies**
   - Run `go mod tidy` to clean up Go dependencies
   - Run `npm audit fix` in frontend/ to update vulnerable packages

## Commands to Run

```bash
# Clean up files
./scripts/cleanup.sh

# Update Go dependencies
go mod tidy

# Update frontend dependencies
cd frontend && npm audit fix

# Remove empty directories
find . -type d -empty -delete 2>/dev/null

# Check for large files that shouldn't be committed
find . -type f -size +1M ! -path "./.git/*" ! -path "./node_modules/*"
```

## Future Maintenance

1. Add pre-commit hooks to prevent:
   - Committing backup files
   - Committing compiled binaries
   - Large file commits

2. Regular cleanup tasks:
   - Run cleanup script before releases
   - Audit dependencies quarterly
   - Review and update documentation

3. Consider adding to Makefile:
   ```makefile
   clean-all: clean clean-deps
       ./scripts/cleanup.sh
   ```