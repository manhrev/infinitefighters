# Task Vô Tận

Game 2D pixel art châm biếm công sở, 1–5 người chơi co-op qua mạng LAN ngay trên trình duyệt. Server Go, client Phaser 3.

## Chạy nhanh

Cần Go, Node.js và npm.

```sh
cd infinitefighters-frontend && npm ci
cd ../infinitefighters-backend && make build && ./infinitefighters
```

Server in ra địa chỉ và mã phòng 4 ký tự. Mở địa chỉ trên trình duyệt, nhập tên và mã phòng. Các cờ server và chế độ dev xem ở [AGENTS.md](AGENTS.md).

## Tài liệu

- [docs/game-scenario.md](docs/game-scenario.md): ý tưởng, bối cảnh, cách chơi, âm thanh
- [docs/levels/](docs/levels/): mỗi màn một file
- [docs/progress.md](docs/progress.md): tiến độ các màn
- [docs/sprite-spec.md](docs/sprite-spec.md), [docs/asset-pipeline.md](docs/asset-pipeline.md): rule asset, cách tạo lại và thêm asset
- [docs/tech-stack.md](docs/tech-stack.md): công nghệ

## Đóng góp

- Cốt truyện mở: màn mới nối tiếp sau màn cuối, được khuyến khích thử lối chơi và góc nhìn mới. Rule chi tiết trong [AGENTS.md](AGENTS.md).
- Sửa luật hoặc số liệu thì cập nhật file màn trong `docs/levels/` và `docs/progress.md` cùng lúc.
- Asset: commit cả nguồn lẫn file runtime; chạy validator trước khi gửi.

## License

[MIT](LICENSE). Áp dụng cho cả code và asset trong repo.
