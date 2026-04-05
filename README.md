# YTDown

YTDown là ứng dụng desktop cho macOS dùng để tải video và trích xuất âm thanh từ các nguồn được `yt-dlp` hỗ trợ. Dự án được xây dựng bằng Go và Wails, tập trung vào một quy trình đơn giản: nhập liên kết, chọn định dạng, theo dõi tiến trình và nhận tệp kết quả ngay trên máy.

## Tính năng chính

- Hỗ trợ tải video chất lượng cao từ nhiều nguồn: **YouTube, Facebook/Instagram Reels, Xiaohongshu (Rednote)**,...
- Tự động nhận diện và xử lý liên kết từ các nền tảng phổ biến
- Hỗ trợ tải từng video hoặc danh sách phát
- Xuất tệp `MP4` hoặc `MP3` (tách nhạc)
- Chọn chất lượng video trước khi tải
- Hiển thị tiến trình tải theo thời gian thực

## Cách dự án xử lý media

YTDown sử dụng hai thành phần chính:

- `yt-dlp`: lấy thông tin video, phân tích liên kết và tải nội dung về máy
- `ffmpeg`: ghép luồng, chuyển đổi định dạng và hỗ trợ xuất `MP3`


## Yêu cầu hệ thống

- macOS 12 trở lên
- Kết nối mạng
- Homebrew để cài `yt-dlp`, `ffmpeg` nếu máy chưa có sẵn

## Cài đặt

### Người dùng cuối

1. Tải bản phát hành `.dmg`
2. Kéo `YTDown.app` vào thư mục `Applications`
3. Mở ứng dụng

Nếu máy chưa có `yt-dlp`, `ffmpeg` hãy cài bằng:

```bash
brew install yt-dlp ffmpeg
```

## Phát triển

### Yêu cầu

- Go 1.22+
- Wails v2 CLI
- Homebrew
- `yt-dlp`
- `ffmpeg`

### Chuẩn bị môi trường

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
brew install yt-dlp
brew install ffmpeg
go mod tidy
```

### Chạy chế độ phát triển

```bash
wails dev
```

### Build ứng dụng

```bash
wails build -platform darwin -tags universal
```

Ứng dụng sau khi build nằm tại:

```text
build/bin/YTDown.app
```

### Tạo bản phát hành

```bash
bash build.sh
```

Script sẽ:

- build `YTDown.app`
- chép `ffmpeg` vào trong ứng dụng
- ký lại app
- tạo thư mục `dist/`
- tạo file `.dmg` nếu hệ thống có `hdiutil`

## Cấu trúc dự án

```text
YTDown/
├── app.go
├── downloader.go
├── compressor.go
├── main.go
├── frontend/
├── scripts/
├── build.sh
└── README.md
```


## License

MIT

## Ủng hộ ly cà phê

Nếu YTDown hữu ích với bạn, bạn có thể ủng hộ mình một ly cà phê:

- MB Bank: `0798888888888`
- Chủ tài khoản: `Nguyen Duc Huy`
