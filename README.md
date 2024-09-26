<img src="winres/reg-finder.ico" align="right" height="70" />

# Windows Registry Finder - GUI



![filter-video-ezgif com-optimize](https://github.com/user-attachments/assets/e1b049fc-667b-4dd6-aa20-a436b10f57f7)



### Features

- [x] Find by keyword
- [x] Filter by Key
- [x] Filter by Type
- [x] Double-clicked to open target registry in `Regedit`


### Build

```
go build -a -ldflags="-s -w -H windowsgui -extldflags '-O2'" .
```

> don't specific build package target to `./main.go` it will make go builder not pick up the `rsrc.syso` so build executable can't run  
