# Windows Registry Finder - GUI


### Features

- [x] Find by keyword
- [ ] Filter by Key
- [ ] Filter by Type
- [x] Double-clicked to open target registry in `Regedit`


### TODO

- Making all features useable
- Improving resource usage and performance
- Improving UI to show value exactly the same we see in `Regedit`


### Build

```
go build -a -ldflags="-s -w -H windowsgui -extldflags '-O2'" .
```