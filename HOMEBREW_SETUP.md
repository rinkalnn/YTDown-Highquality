# Homebrew Cask Setup Guide

## For Maintainers: Hosting Your Own Tap

To allow users to install YTDown via Homebrew, you have several options:

### Option 1: Create a Personal Homebrew Tap (Recommended)

A Homebrew Tap is essentially a Git repository containing cask formulas.

1. **Create a new GitHub repository** named `homebrew-ytdown`

2. **Add the formula file** to your tap:
   ```
   Formula/ytdown.rb
   ```

3. **Push to GitHub** and make it public

4. **Users can then install with:**
   ```bash
   brew tap JustinNguyen9979/ytdown
   brew install --cask ytdown
   ```

5. **Users can update with:**
   ```bash
   brew upgrade --cask ytdown
   # or
   brew update && brew upgrade --cask ytdown
   ```

### Option 2: Official Homebrew Cask

To submit to the official Homebrew Cask repository:
1. Fork https://github.com/Homebrew/homebrew-cask
2. Add your formula to the `Casks/` directory
3. Create a pull request

This allows users to install with:
```bash
brew install --cask ytdown
```

## How It Works

The Homebrew formula (`Formula/ytdown.rb`):
- Automatically fetches the latest DMG from your GitHub releases
- Uses the GitHub API to get the download URL
- Supports auto-updates via the appcast (Atom feed)
- Works with the auto-update mechanism in your app

## User Installation Instructions

### First Install
```bash
# If using a personal tap
brew tap JustinNguyen9979/ytdown
brew install --cask ytdown

# If it's official homebrew-cask (no tap needed)
brew install --cask ytdown
```

### Update
```bash
# Check for updates
brew outdated --cask

# Update YTDown
brew upgrade --cask ytdown

# Or let the app update itself through the built-in update mechanism
```

## CI/CD Integration

The GitHub Actions workflow now:
1. Builds the app automatically on every push to `main`
2. Creates releases with the format `YYYY.MM.DD` (e.g., `2026.04.12`)
3. Uploads DMG and ZIP files to releases
4. Homebrew automatically picks up new releases

No Apple signing or notarization needed!

## Environment Variables (No longer needed but removed)

Previously used for Apple signing:
- `APPLE_CERTIFICATE_BASE64` - ❌ Removed
- `APPLE_CERTIFICATE_PASSWORD` - ❌ Removed
- `KEYCHAIN_PASSWORD` - ❌ Removed
- `APPLE_ID` - ❌ Removed
- `APPLE_APP_PASSWORD` - ❌ Removed
- `APPLE_TEAM_ID` - ❌ Removed
- `APPLE_SIGNING_IDENTITY` - ❌ Removed

You can safely delete these from your GitHub secrets.

## Testing the Setup

1. **Test locally:**
   ```bash
   # Install from your local formula
   brew install --cask ./Formula/ytdown.rb
   
   # Or after pushing to tap
   brew tap JustinNguyen9979/ytdown
   brew install --cask ytdown
   ```

2. **Check app update mechanism:**
   - Open the app
   - Check for updates in the menu
   - The app should detect the latest GitHub release automatically

3. **Test Homebrew update:**
   ```bash
   brew upgrade --cask ytdown
   ```
