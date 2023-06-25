Folder Structure Convention and Packages
===========================
```bash
├── main
│   ├── main.go `package: main`
│   ├── backend
│   │   ├── listener.go
│   ├── config
│   │   ├── config.go
│   ├── storage
│   │   ├── mongodb.go
│   │   ├── objects.go
│   ├── tests
│   │   ├── main_test.go
```

Framework Dependencies
===========================

- ws https://github.com/gobwas/ws 

Tests
===========================
- `/tests/main_test.go`