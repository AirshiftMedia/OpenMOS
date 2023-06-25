Folder Structure Convention and Packages
===========================
```bash
├── main
│   ├── main.go [package: main]
│   ├── backend
│   │   ├── listener.go [package: backend]
│   ├── config
│   │   ├── config.go [package: config]
│   ├── storage
│   │   ├── mongodb.go [package: storage]
│   │   ├── objects.go [package: storage]
│   ├── tests
│   │   ├── main_test.go [package: tests]
```

Framework Dependencies
===========================

- ws https://github.com/gobwas/ws 

Tests
===========================
- `/tests/main_test.go`