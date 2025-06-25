# Homebrew Formula Migration Guide

This document explains the migration from a separate Homebrew tap repository to having the formula integrated directly in the main smtp-edc repository.

## What Changed

### Before (Separate Tap)
- Formula lived in `asachs01/homebrew-smtp-edc`
- Required manual updates via GitHub Actions triggers
- Installation: `brew tap asachs01/smtp-edc && brew install smtp-edc`

### After (Integrated Formula)
- Formula lives in `Formula/smtp-edc.rb` in the main repository
- Automatically updated by GoReleaser on each release
- Installation: `brew tap asachs01/smtp-edc && brew install smtp-edc`

## Benefits

1. **Single Repository**: Everything lives in one place
2. **Automated Updates**: GoReleaser handles formula updates automatically
3. **Consistency**: No manual synchronization needed between repositories
4. **Simplified CI/CD**: One workflow handles everything

## For Users

The installation process remains the same:

```bash
# Add the tap
brew tap asachs01/smtp-edc

# Install smtp-edc
brew install smtp-edc
```

## For Development

### Testing Formula Locally

Use the provided test script:

```bash
./scripts/test_homebrew.sh
```

### Release Process

1. Tag a new version: `git tag v1.2.0`
2. Push the tag: `git push origin v1.2.0`
3. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Update the Homebrew formula
   - Commit the updated formula to the repository

## Migration Steps Completed

- [x] Added `brews` configuration to `.goreleaser.yml`
- [x] Created `Formula/smtp-edc.rb` in main repository
- [x] Updated GitHub Actions workflow to use GoReleaser
- [x] Removed dependency on separate tap repository
- [x] Updated README with new installation instructions
- [x] Created test script for local validation

## Cleanup

After confirming the new setup works:

1. Archive the old `asachs01/homebrew-smtp-edc` repository
2. Remove any `HOMEBREW_TOKEN` secrets that were used for the old workflow
3. Update any documentation that references the old installation method
