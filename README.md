<img src="winres/reg-finder.ico" align="right" height="70" />

# Windows Registry Finder - GUI


![reg-finder-ezgif com-optimize](https://github.com/user-attachments/assets/d8cfd0a3-bce5-4148-861f-85eac662701b)



### Features

- [x] Find by keyword
- [x] Filter by Key
- [x] Filter by Type
- [x] Double-clicked to open target registry in `Regedit`


### TODO

- Improving resource usage and performance
- Improving UI to show value exactly the same we see in `Regedit`


### Build

```
go build -a -ldflags="-s -w -H windowsgui -extldflags '-O2'" .
```

> don't specific build package target to `./main.go` it will make go builder not pick up the `rsrc.syso` so build executable can't run  
