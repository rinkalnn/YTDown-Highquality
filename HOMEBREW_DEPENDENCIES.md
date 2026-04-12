# Homebrew Dependencies Setup

## Overview

YTDown now uses **Homebrew** to manage its dependencies instead of bundling them. This approach has several advantages:

✅ **Always up-to-date** - Tools auto-update when you run `brew upgrade`
✅ **Smaller app size** - No bundled binaries in the DMG
✅ **Better performance** - System binaries are optimized for your machine
✅ **Easy updates** - Single command: `brew upgrade --cask ytdown`

## Required Dependencies

YTDown requires 3 main tools via Homebrew:

1. **ffmpeg** - Video processing & extraction
2. **yt-dlp** - YouTube/video downloading engine
3. **gallery-dl** - Image gallery downloading engine

## Installation

### Automatic (On First Launch)

When you first open YTDown:

1. App checks if dependencies are installed
2. If missing → **Dialog appears** asking to install
3. Click "Install" → app runs: `brew install ffmpeg yt-dlp gallery-dl`
4. App restarts automatically
5. You're ready to use!

### Manual Installation

If you prefer to install manually:

```bash
brew install ffmpeg yt-dlp gallery-dl
```

Or install individually:

```bash
brew install ffmpeg
brew install yt-dlp
brew install gallery-dl
```

### Verify Installation

```bash
# Check all tools are installed
brew list | grep -E "ffmpeg|yt-dlp|gallery-dl"

# Check versions
ffmpeg -version
yt-dlp --version
gallery-dl --version
```

## Updating

To update dependencies to the latest versions:

```bash
# Update all Homebrew packages
brew upgrade

# Or update specific tools
brew upgrade ffmpeg yt-dlp gallery-dl
```

This also updates YTDown itself via Homebrew Cask:

```bash
brew upgrade --cask ytdown
```

## Troubleshooting

### "Tool not found in PATH"

If you're getting "not found in PATH", try:

```bash
# Reinstall the tool
brew install --force ffmpeg

# Or check Homebrew installation
brew doctor
```

### I've installed Homebrew but it's not in PATH

Homebrew might not be in your PATH. Run:

```bash
/usr/local/bin/brew --version  # Intel Mac
# or
/opt/homebrew/bin/brew --version  # Apple Silicon Mac
```

Then add it to your PATH in `~/.zprofile`:

```bash
export PATH="/opt/homebrew/bin:$PATH"
```

### Permission Denied When Installing

If `brew install` asks for a password, that's normal. Enter your Mac password when prompted.

### Upgrading from Old Version

If you had an older version of YTDown with bundled binaries:

```bash
# Uninstall old version
brew uninstall --cask ytdown

# Tap the new version
brew tap JustinNguyen9979/ytdown

# Install new version
brew install --cask ytdown

# Install dependencies
brew install ffmpeg yt-dlp gallery-dl
```

## System Requirements

- **macOS**: Big Sur (11.0) or newer
- **Homebrew**: Must be installed (https://brew.sh)
- **Disk Space**: ~500MB for all dependencies

## Architecture

```
YTDown App (DMG)
    ↓
On Startup
    ├─ Check: ffmpeg in PATH?
    ├─ Check: yt-dlp in PATH?
    └─ Check: gallery-dl in PATH?
         ↓
    If missing → Show Dialog
         ↓
    User Click "Install"
         ↓
    App runs: brew install ffmpeg yt-dlp gallery-dl
         ↓
    Dependencies Available
         ↓
    App Ready to Use
```

## For Developers

### Build Requirements

To build YTDown locally:

```bash
brew install go node ffmpeg yt-dlp gallery-dl
cd YTDown
make build
```

### CI/CD

GitHub Actions automatically:
- Builds without dependencies (no bundling)
- Creates DMG release
- Updates Homebrew tap
- Users get the new version via `brew upgrade --cask ytdown`

## FAQ

**Q: Can I use older versions of ffmpeg/yt-dlp?**
A: Not recommended. Homebrew always installs the latest stable version. If you need a specific version, you can build from source.

**Q: What if I don't want to use Homebrew?**
A: You can compile from source or use the old bundled version. Homebrew is recommended for easy updates.

**Q: Do I need administrator password?**
A: Only when installing via `brew install`. Your password is your Mac login password.

**Q: Can I uninstall dependencies?**
A: Yes, but YTDown won't work:
```bash
brew uninstall ffmpeg yt-dlp gallery-dl
```

**Q: What if Homebrew is not installed?**
A: YTDown will show an error on first launch with a link to install Homebrew (https://brew.sh)

## See Also

- [Homebrew Documentation](https://docs.brew.sh)
- [yt-dlp GitHub](https://github.com/yt-dlp/yt-dlp)
- [gallery-dl GitHub](https://github.com/mikf/gallery-dl)
- [FFmpeg Documentation](https://ffmpeg.org/documentation.html)
