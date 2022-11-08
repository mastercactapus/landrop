# LAN Drop

A simple HTTP server for sharing files on your local network.

## Usage

To share the current directory (read-only) at `http://landrop.local:8000` run:

```bash
go run github.com/mastercactapus/landrop
```

To allow uploads, run:

```bash
go run github.com/mastercactapus/landrop -w
```
