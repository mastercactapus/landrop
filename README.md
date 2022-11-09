# LAN Drop

A simple HTTP server for sharing files on your local network.

![image](https://user-images.githubusercontent.com/595010/200705887-eb0a771d-4c1b-45a0-b3d7-f689e99a8ed3.png)

## Usage

To share the current directory (read-only) at `http://landrop.local:8000` run `landrop` from the directory you want to make available.

To run/install directly from source:

```bash
go run github.com/mastercactapus/landrop
```

To allow uploads, add the `-w` flag:

```bash
go run github.com/mastercactapus/landrop -w
```
