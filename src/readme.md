Source File Structure and Packages
===========================
```bash
├── main
│   ├── main.go [package: main]
│   │
│   ├── handlers
│   │   ├── listener.go [package: handlers]
│   ├── observer
│   │   ├── sentry.go [package: observer]
│   ├── models
│   │   ├── objects.go [package: models]
│   ├── utils
│   │   ├── config.go [package: utils]
│   ├── storage
│   │   ├── storage.go [package: storage]
```

Framework Dependencies
===========================

WebSocket is based on Gorilla framework. While the project is currently abandoned, it is still expected to be a valid option.

- github.com/gorilla/websocket

Tests
===========================
- `/tests/main_test.go`