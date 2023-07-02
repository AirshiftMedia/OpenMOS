Source File Structure and Packages
===========================
```bash
├── main
│   ├── main.go [package: main]
│   │
│   ├── handlers
│   │   ├── listener.go [package: handlers]
│   ├── config
│   │   ├── config.go [package: config]
│   ├── observer
│   │   ├── sentry.go [package: observer]
│   ├── models
│   │   ├── model.go [package: model]
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