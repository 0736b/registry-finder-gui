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


```
go get github.com/akavel/rsrc
rsrc -manifest .\registry-finder-gui.manifest -o rsrc.syso

go build -ldflags "-extldflags=-static -H windowsgui" .
```