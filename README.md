# YTDown

YTDown là ứng dụng tải video YouTube cho macOS, được xây dựng bằng Wails v2 và Go. App hỗ trợ tải video đơn lẻ, tải hàng loạt, lấy danh sách video từ playlist, trích xuất MP3, và nén media ngay trong giao diện desktop.

Điểm đặc biệt của YTDown là toàn bộ ứng dụng được xây dựng với sự hỗ trợ xuyên suốt của AI. Từ ý tưởng, cấu trúc, logic xử lý cho đến quá trình hoàn thiện sản phẩm, dự án này được thực hiện với sự đồng hành của các mô hình như Claude Opus 4-6, ChatGPT 5.4 và Gemini 3 Flash. Đây không chỉ là một ứng dụng tiện ích, mà còn là một thử nghiệm thực tế về cách AI có thể trở thành cộng sự đúng nghĩa trong quá trình phát triển phần mềm.

## Tính năng

- Tải một hoặc nhiều link YouTube trong cùng một phiên.
- Hỗ trợ playlist và tự lấy danh sách video trong playlist.
- Xuất ra MP4 hoặc MP3.
- Chọn chất lượng tải từ 360p đến Best Quality.
- Theo dõi tiến trình tải theo thời gian thực.
- Tích hợp nén file media bằng `ffmpeg`.

## Cách app xử lý phụ thuộc

- `ffmpeg` được bundle sẵn bên trong `YTDown.app`.
- Khi chạy, app sẽ kiểm tra `yt-dlp` trong hệ thống.
- Nếu máy đã có `yt-dlp` qua Homebrew, app dùng ngay bản đó.
- Nếu máy chưa có `yt-dlp` nhưng đã có Homebrew, app sẽ thử tự chạy `brew install yt-dlp`.
- Nếu máy chưa có Homebrew hoặc quá trình cài thất bại, app sẽ hiện thông báo để user tự cài.

## Yêu cầu hệ thống

- macOS 12 trở lên.
- Kết nối mạng để tải video và để Homebrew cài `yt-dlp` khi cần.
- Homebrew là phương án chuẩn để cài `yt-dlp`.

## Hướng dẫn cài cho người dùng cuối

### Cách 1: Dùng bản `.dmg`

1. Tải file `.dmg` phát hành.
2. Mở file `.dmg`.
3. Kéo `YTDown.app` vào `Applications`.
4. Mở app từ `Applications`.

Ở lần chạy đầu:

- Nếu máy đã có Homebrew và `yt-dlp`, app sẽ hoạt động ngay.
- Nếu máy có Homebrew nhưng chưa có `yt-dlp`, app sẽ tự thử cài `yt-dlp`.
- Nếu máy chưa có Homebrew, app sẽ báo lỗi hướng dẫn cài Homebrew trước.

### Cách 2: Cài Homebrew và `yt-dlp` thủ công trước khi mở app

Nếu muốn chủ động cài sẵn phụ thuộc, chạy:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew install ffmpeg
brew install yt-dlp
```

Sau đó mở lại `YTDown.app`.

## Khi app báo lỗi thiếu Homebrew hoặc `yt-dlp`

### Trường hợp 1: Chưa có Homebrew

Chạy lệnh sau trong Terminal:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Sau khi cài xong, chạy tiếp:

```bash
brew install ffmpeg
brew install yt-dlp
```

Rồi mở lại app.

### Trường hợp 2: Có Homebrew nhưng app không cài được `yt-dlp`

Chạy thủ công:

```bash
brew install ffmpeg
brew install yt-dlp
```

Nếu đã cài rồi mà vẫn lỗi:

```bash
brew upgrade yt-dlp
which yt-dlp
yt-dlp --version
```

## Build từ source cho developer

### Yêu cầu

- Go 1.22+
- Wails CLI v2
- Homebrew
- `yt-dlp`

### Cài môi trường

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
brew install yt-dlp
go mod tidy
```

### Chạy dev mode

```bash
wails dev
```

### Build app

```bash
wails build -platform darwin -tags universal
```

App output sẽ nằm ở:

```text
build/bin/YTDown.app
```

## Build bản phát hành `.dmg`

Repo có sẵn script build:

```bash
FFMPEG_PATH=/absolute/path/to/ffmpeg bash build.sh
```

Script này sẽ:

- build `YTDown.app`
- bundle `ffmpeg` vào `YTDown.app/Contents/Resources`
- ký lại app bằng ad-hoc signature
- tạo `dist/YTDown.app`
- tạo `dist/YTDown-1.0.0.dmg`
- thêm shortcut `Applications` vào DMG để user kéo thả trực tiếp

## Lưu ý phát hành

- App hiện phụ thuộc vào Homebrew để cài `yt-dlp`.
- Điều này phù hợp với máy đã có Homebrew hoặc môi trường cá nhân của bạn.
- Nếu muốn phát hành cho user hoàn toàn không biết kỹ thuật, bạn nên chuẩn bị thêm một luồng onboarding rõ ràng hoặc installer riêng cho Homebrew.
- `ffmpeg` đang được bundle theo bản bạn chỉ định lúc build.

## Cấu trúc chính của dự án

```text
YTDown/
├── main.go
├── app.go
├── downloader.go
├── compressor.go
├── frontend/
├── build/
├── scripts/
├── build.sh
└── README.md
```

## Khắc phục sự cố nhanh

- App mở lên nhưng báo thiếu Homebrew:
  Cài Homebrew từ `https://brew.sh`, rồi mở lại app.
- App báo cài `yt-dlp` thất bại:
  Chạy `brew install yt-dlp` thủ công trong Terminal.
- App tải MP3 lỗi:
  Kiểm tra lại `ffmpeg` trong bản build hoặc build lại bằng `FFMPEG_PATH`.
- Tốc độ tải chậm:
  Kiểm tra version `yt-dlp` bằng `yt-dlp --version`. Trên máy này, bản Homebrew đang cho kết quả nhanh hơn bản standalone đã test trước đó.

## License

MIT

## Mời mình ly cà phê

Nếu app hữu ích với bạn, bạn có thể ủng hộ mình một ly cà phê:

- MB Bank: `0798888888888`
- Chủ tài khoản: `Nguyen Duc Huy`

Mỗi sự ủng hộ nhỏ đều là động lực để mình tiếp tục dành thời gian cải thiện app, sửa lỗi và thêm tính năng mới. Cảm ơn bạn rất nhiều.
