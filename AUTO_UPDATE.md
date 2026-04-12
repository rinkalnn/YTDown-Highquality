# Auto Update

Repo này đã có khung auto-release và app-side updater cho macOS.

## Versioning

- Version được sinh tự động theo múi giờ `Asia/Ho_Chi_Minh`
- Format: `yyyy.mm.dd`
- Workflow release chạy khi push lên `main`

## Luồng hoạt động

1. Push code lên `main`
2. GitHub Actions tự tính version theo ngày
3. Chạy `build.sh`
4. Tạo release GitHub với tag đúng version ngày đó
5. Upload:
   - `YTDown-<version>.dmg`
   - `YTDown-<version>.zip`
6. App khi mở sẽ gọi GitHub Releases API
7. Nếu có version mới hơn, app hiện banner update
8. User bấm update, app tải `.dmg`, thay app bundle, rồi mở lại app

## File chính

- `.github/workflows/release.yml`
- `build.sh`
- `app_update.go`
- `frontend/app.js`

## GitHub Secrets

Từ **2026.04.12 trở đi**, workflow không cần secret Apple nữa:

✅ **Ad-hoc signing** - được xử lý tự động (không cần secret)
❌ **Apple code signing** - đã bỏ
❌ **Notarization** - đã bỏ

Nếu bạn vẫn có secrets Apple trong GitHub, có thể xóa chúng để clean:
- `APPLE_CERTIFICATE_BASE64`
- `APPLE_CERTIFICATE_PASSWORD`
- `KEYCHAIN_PASSWORD`
- `APPLE_SIGNING_IDENTITY`
- `APPLE_ID`
- `APPLE_APP_PASSWORD`
- `APPLE_TEAM_ID`
- Base64 file đó rồi đưa vào secret

`APPLE_CERTIFICATE_PASSWORD`
- Password của file `.p12`

`KEYCHAIN_PASSWORD`
- Mật khẩu tạm cho keychain trên runner CI

`APPLE_SIGNING_IDENTITY`
- Ví dụ: `Developer ID Application: Your Name (TEAMID)`

`APPLE_ID`
- Apple ID dùng cho notarization

`APPLE_APP_PASSWORD`
- App-specific password của Apple ID

`APPLE_TEAM_ID`
- Team ID trong Apple Developer

## Lưu ý thực tế

- Nếu không có Apple signing secrets:
  - workflow vẫn build release được
  - nhưng app sẽ dùng ad-hoc signing
- Nếu có đầy đủ secrets:
  - workflow sẽ import certificate
  - sign `.app` và `.dmg`
  - notarize `.dmg`
  - staple notarization ticket

## Giới hạn hiện tại

- Updater hiện check release latest từ GitHub repo
- Chưa có delta update
- Chưa có appcast riêng kiểu Sparkle
- Luồng hiện tại là tải full `.dmg` rồi thay app bundle

## Khuyến nghị

Nếu sau này muốn updater chuẩn macOS hơn nữa:
- chuyển sang Sparkle
- dùng appcast feed cố định
- giữ GitHub Actions chỉ làm nhiệm vụ build/sign/notarize/publish artifact
