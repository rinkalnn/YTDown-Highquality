CI/CD Setup - Summary of Changes
================================

✅ **Completed**: Your CI/CD workflow is now fully configured!

## What Was Changed

### 1. GitHub Actions Workflow (`.github/workflows/release.yml`)
**Removed:**
- ❌ Apple certificate import step
- ❌ Notarization step
- ❌ All Apple signing environment variables

**Kept:**
- ✅ Automatic build on push to `main`
- ✅ Auto version generation (`YYYY.MM.DD` format)
- ✅ Release creation with DMG & ZIP files
- ✅ Ad-hoc code signing (automatic)

### 2. Wails Configuration (`wails.json`)
**Cleaned up:**
- ❌ Removed empty `codeSigningIdentity`, `codesigningidentity`, `codesignidentity`, `signingidentity` fields
- ✅ Kept `signoptions: "nodot,adhoc"` for ad-hoc signing

### 3. Build Script (`build.sh`)
**Status:** ✅ No changes needed - already using ad-hoc signing

### 4. New: Homebrew Cask Formula
**Created:** `Formula/ytdown.rb`
- Automatically fetches the latest DMG from GitHub releases
- Supports auto-updates via Atom feed
- Allows users to do `brew upgrade --cask ytdown`

### 5. Documentation
**Created:** 
- `HOMEBREW_SETUP.md` - Complete guide for Homebrew Cask setup
- Updated `AUTO_UPDATE.md` - Clarified that Apple signing is no longer needed

## How It Works Now

### For You (Developer)
1. Write code and push to `main`
2. GitHub Actions automatically:
   - ✅ Builds the app with ad-hoc signing
   - ✅ Creates a release tagged with today's date
   - ✅ Uploads DMG and ZIP files
3. Done! No secrets or certificates needed.

### For Users

#### Option A: Direct Download
Users download DMG from: `https://github.com/JustinNguyen9979/YTDown/releases`

#### Option B: In-App Updates
Users open the app → "Check for Updates" → App automatically detects new versions and prompts to update

#### Option C: Homebrew Cask (Recommended)
1. First, you need to create a Homebrew tap:
   ```bash
   # Create a new repo: homebrew-ytdown
   # Put Formula/ytdown.rb in it
   ```

2. Then users can:
   ```bash
   brew tap JustinNguyen9979/ytdown
   brew install --cask ytdown
   brew upgrade --cask ytdown  # to update
   ```

## Next Steps

### Immediate (1-2 minutes)
1. ✅ Review the changes above (all done!)
2. ✅ Read `HOMEBREW_SETUP.md` for Homebrew integration

### Soon (Optional but Recommended)
1. **Create Homebrew Tap** (if you want brew support):
   - Create a new GitHub repo named `homebrew-ytdown`
   - Copy the `Formula/ytdown.rb` file to the root
   - Push to GitHub
   - Users can then use Homebrew to install

2. **Clean up GitHub Secrets** (if you had them):
   - Delete these if they exist:
     - `APPLE_CERTIFICATE_BASE64`
     - `APPLE_CERTIFICATE_PASSWORD`
     - `KEYCHAIN_PASSWORD`
     - `APPLE_SIGNING_IDENTITY`
     - `APPLE_ID`
     - `APPLE_APP_PASSWORD`
     - `APPLE_TEAM_ID`

### Test the Setup
```bash
# Push a small change to main
git add .
git commit -m "CI/CD: Remove Apple signing, add Homebrew support"
git push origin main

# Wait for GitHub Actions to run
# Check release at: https://github.com/JustinNguyen9979/YTDown/releases

# Test local build
bash build.sh

# Check the DMG was created
ls -lh dist/YTDown-*.dmg
```

## Version Format

Every release is tagged with today's date in `YYYY.MM.DD` format:
- `2026.04.12` (April 12, 2026)
- `2026.04.13` (April 13, 2026)
- etc.

This makes it easy to identify when each release was built and automatically compare versions (newer date = newer version).

## Architecture

```
Push to main
    ↓
GitHub Actions Workflow
    ├─ Setup Go & Node
    ├─ Build app with ad-hoc signing
    ├─ Create DMG & ZIP
    └─ Create GitHub Release + Tag
         ↓
    GitHub Releases (with DMG asset)
         ↓
    ┌────┴────────┐
    ↓             ↓
In-App Update    Homebrew Cask
(checks API)     (fetches DMG)
```

## Troubleshooting

**Issue**: Workflow fails
- Check GitHub Actions logs
- Ensure Go & Node versions in `go.mod` and workflow match

**Issue**: Release not created
- Check workflow output for errors
- Verify local `bash build.sh` works first

**Issue**: App doesn't detect new version
- Verify DMG is uploaded to release
- Check app's `app_update.go` - looks for `.dmg` files

**Issue**: Homebrew installation fails
- Ensure `homebrew-ytdown` tap is public
- Verify `Formula/ytdown.rb` is in root directory
- Test with: `brew install --cask --verbose homebrew-ytdown/ytdown`

## Questions?

See:
- `HOMEBREW_SETUP.md` - Homebrew Cask setup & usage
- `AUTO_UPDATE.md` - Auto-update mechanism details
- `.github/workflows/release.yml` - Workflow configuration
