# LAN Drop

A simple HTTP server for sharing files on your local network.

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
