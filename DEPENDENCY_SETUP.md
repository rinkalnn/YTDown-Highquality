#!/bin/bash

# YTDown - Auto-Dependency Installation Setup
# Summary of all changes made

cat << 'EOF'

🎯 SETUP COMPLETE: Homebrew Auto-Dependency Installation
==========================================================

## What Changed?

### ✅ Backend Changes

1. **dependency_checker.go** (NEW)
   - CheckDependencies() - Check if ffmpeg, yt-dlp, gallery-dl are installed
   - InstallDependencies() - Auto-install via `brew install`
   - PromptToInstallDependencies() - Show user dialog to agree

2. **app.go** (UPDATED)
   - Updated startup() - Check dependencies on app launch
   - Updated CheckBinaries() - Use new dependency system

3. **build.sh** (UPDATED)
   - Removed FFmpeg bundling script
   - No longer requires ffmpeg in GitHub Actions

### ✅ Frontend Changes

1. **app.js** (UPDATED)
   - Added event listener for 'dependencies-missing'
   - Shows confirmation dialog when dependencies missing
   - Calls backend to install when user agrees

### ✅ CI/CD Changes

1. **.github/workflows/release.yml** (UPDATED)
   - Removed: `brew install ffmpeg` step
   - Build now doesn't require ffmpeg in CI

### ✅ Configuration

1. **wails.json** (ALREADY UPDATED)
   - Using ad-hoc signing (no Apple certificates needed)

### ✅ Documentation

1. **HOMEBREW_DEPENDENCIES.md** (NEW)
   - Complete guide on the new dependency system
   - Installation instructions (auto + manual)
   - Troubleshooting guide

2. **README.md** (UPDATED)
   - Added Homebrew Cask installation instructions
   - Updated downloads section
   - Updated development setup

3. **HOMEBREW_SETUP.md** (EXISTING)
   - Info on creating Homebrew tap

4. **CI_CD_SETUP.md** (EXISTING)
   - Summary of CI/CD changes

## How It Works

### For Users

1. Download DMG and install YTDown
2. Launch app for first time
3. App checks: "Is ffmpeg installed? Is yt-dlp installed? Is gallery-dl installed?"
4. If any missing → Dialog appears:
   ```
   ⚠️ Missing Dependencies
   
   YTDown requires:
   • ffmpeg
   • yt-dlp
   • gallery-dl
   
   Install now via Homebrew?
   [Yes] [No]
   ```
5. If Yes → App runs: `brew install ffmpeg yt-dlp gallery-dl`
6. After install → App ready to use

### For Development (Local Build)

```bash
# Clone repo
git clone https://github.com/JustinNguyen9979/YTDown.git
cd YTDown

# Install dev tools
brew install go node

# (Optional) Pre-install dependencies to avoid prompt
brew install ffmpeg yt-dlp gallery-dl

# Run development
wails dev

# Build
bash build.sh
```

### For CI/CD (GitHub Actions)

1. Push code to main
2. GitHub Actions runs:
   - Setup Go & Node
   - Build app (no ffmpeg needed)
   - Create release with DMG & ZIP
3. publish-homebrew.yml triggers:
   - Updates homebrew-ytdown repo
   - Users can: `brew upgrade --cask ytdown`

## Testing the Setup

### Local Test
```bash
# Build locally
bash build.sh

# Should complete without errors
# (no ffmpeg bundling anymore)

# Check DMG was created
ls -lh dist/YTDown-*.dmg
```

### Test Dependency Check
1. Open YTDown
2. Should NOT show any warning (if ffmpeg/yt-dlp/gallery-dl installed)
3. OR if uninstalling one tool:
   ```bash
   brew uninstall ffmpeg
   ```
   Then reopen app → should show dependency dialog

### Test Installation
1. Have one tool missing (e.g., uninstall gallery-dl)
2. Open app
3. Dialog appears
4. Click "Install"
5. App runs `brew install gallery-dl`
6. After completion, app should be ready

## Architecture Diagram

```
User Action
    ↓
┌─────────────────────┐
│  Download DMG       │
│  Install YTDown.app │
└─────────────────────┘
    ↓
Open App
    ↓
┌─────────────────────────┐
│ app.startup()           │
│ CheckDependencies()     │
└─────────────────────────┘
    ↓
Check: ffmpeg in PATH? ────→ Found ↓
Check: yt-dlp in PATH?  ────→ Found ↓  Ready to Use ✅
Check: gallery-dl in PATH? ──→ Found ↓
    ↓
   ANY Missing?
    ↓
   YES
    ↓
Emit: 'dependencies-missing' event
    ↓
Frontend receives event
    ↓
┌──────────────────────────┐
│ Show Dialog to User:     │
│ "Install dependencies?" │
└──────────────────────────┘
    ↓
User clicks "Install"
    ↓
PromptToInstallDependencies()
    ↓
Run: brew install ffmpeg yt-dlp gallery-dl
    ↓
Success?
    ↓
YES ↓              NO ↓
Show "✅ Done"   Show "❌ Failed"
Required Restart  Try Manual Install
```

## Files Modified

1. ✅ `.github/workflows/release.yml` - Removed ffmpeg install
2. ✅ `.github/workflows/publish-homebrew.yml` - Already created
3. ✅ `app.go` - Updated startup & CheckBinaries()
4. ✅ `app.js` - Added dependencies-missing handler
5. ✅ `build.sh` - Removed ffmpeg bundling
6. ✅ `wails.json` - Already cleaned
7. ✅ `README.md` - Updated instructions

## Files Created

1. ✅ `dependency_checker.go` - Core dependency logic
2. ✅ `HOMEBREW_DEPENDENCIES.md` - User documentation
3. ✅ `CI_CD_SETUP.md` - CI/CD documentation
4. ✅ `Formula/ytdown.rb` - Homebrew cask formula

## Next Steps

1. Commit changes:
   ```bash
   git add .
   git commit -m "feat: Auto-install dependencies via Homebrew

   - Add dependency_checker.go for automatic detection
   - No longer bundle ffmpeg (users install via brew)
   - App prompts to install missing tools on startup
   - Update CI/CD workflow - remove ffmpeg dependency
   - Users: brew tap JustinNguyen9979/ytdown && brew install --cask ytdown"
   ```

2. Push to main:
   ```bash
   git push origin main
   ```

3. Wait for GitHub Actions:
   - release.yml builds the app
   - publish-homebrew.yml updates the tap repo
   - Users can install via Homebrew

4. Test:
   ```bash
   # Remove a tool to test dialog
   brew uninstall gallery-dl
   
   # Open app - should show dialog
   open dist/YTDown.app
   
   # Click Install - should auto-install
   ```

## Advantages of This Approach

✅ **App stays small** - No bundled binaries
✅ **Tools always up-to-date** - brew upgrade handles it
✅ **Better performance** - System optimized binaries
✅ **Easier maintenance** - No need to update bundled versions
✅ **Better UX** - Auto-install on first run, then forget about it
✅ **No Apple signing needed** - Ad-hoc signing only
✅ **Users can upgrade with one command** - `brew upgrade --cask ytdown`

## Questions?

See:
- `HOMEBREW_DEPENDENCIES.md` - User guide
- `CI_CD_SETUP.md` - CI/CD details
- `HOMEBREW_SETUP.md` - Tap setup guide
- `.github/workflows/` - Workflow definitions

EOF
