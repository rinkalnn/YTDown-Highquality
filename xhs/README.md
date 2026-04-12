# XHS Module

Package `xhs` dùng để resolve short-link Xiaohongshu, tải HTML note, parse `__INITIAL_STATE__`, rồi trả ra URL ảnh hoặc video đã sẵn sàng để download.

## API hiện có

API nên dùng:

```go
info, err := xhs.FetchWithOptions(ctx, noteURL, xhs.FetchOptions{
    Cookie:      "web_session=YOUR_COOKIE",
    UserAgent:   "Mozilla/5.0 ...",
    ImageFormat: "best",
})
```

API tương thích ngược:

```go
info, err := xhs.Fetch(ctx, noteURL, "web_session=YOUR_COOKIE", "png")
```

`Fetch()` chỉ là wrapper mỏng quanh `FetchWithOptions()`.

## Tham số `FetchOptions`

`Cookie`
- Toàn bộ header `Cookie`.
- Thực tế thường truyền `web_session=...`.
- Được dùng ở cả bước resolve short-link và tải HTML note.

`UserAgent`
- Override `User-Agent`.
- Nếu bỏ trống, module dùng UA mặc định dạng desktop browser.

`Referer`
- Override `Referer`.
- Nếu bỏ trống, mặc định là `https://www.xiaohongshu.com/`.

`AcceptLanguage`
- Override `Accept-Language`.
- Nếu bỏ trống, mặc định là `zh-CN,zh;q=0.9,en;q=0.8`.

`ImageFormat`
- Điều khiển URL ảnh trả về.
- Giá trị hỗ trợ: `""`, `best`, `auto`, `png`, `webp`, `jpeg`, `heic`.
- `""`, `best`, `auto`: trả ảnh gốc để ưu tiên chất lượng cao nhất.
- `png`, `webp`, `jpeg`, `heic`: trả URL CDN có ép format.

`Timeout`
- Kiểu `time.Duration`.
- Áp dụng cho cả bước resolve short-link lẫn bước tải HTML.
- Nếu `<= 0`, mặc định là `30 * time.Second`.

## Dữ liệu trả về

`NoteInfo`
- `NoteID`: ID note sau khi resolve.
- `Title`: tiêu đề bài viết.
- `Author`: nickname tác giả.
- `Type`: `image` hoặc `video`.
- `ImageURLs`: danh sách URL ảnh hoặc live-photo stream nếu là bài ảnh.
- `VideoURL`: URL video nếu là bài video.

## Hành vi hiện tại

- Tự resolve `xhslink.com` sang URL note thật trước khi parse.
- Với bài ảnh, mặc định ưu tiên ảnh gốc để tránh mất nét.
- Với bài video, hiện trả `masterUrl` đầu tiên tìm được.

## Chưa hỗ trợ riêng

- Chọn quality level riêng cho video.
- Chọn codec/container riêng cho video.
- Retry policy hoặc proxy config.
