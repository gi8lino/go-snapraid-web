# go-snapraid-webui

A Web UI dashboard for [go-snapraid](https://github.com/gi8lino/go-snapraid.git), built in Go. It provides a responsive, filterable, and sortable view of SnapRAID operations such as `diff`, `sync`, `scrub`, `smart`, and `touch`.

## Usage

```bash
go-snapraid-webui [flags]
```

## Flags

| Flag               | Short | Default   | Description                              |
| ------------------ | ----- | --------- | ---------------------------------------- |
| `--listen-address` | `-a`  | `:8080`   | Address to listen on (e.g., `:8080`)     |
| `--output-dir`     | `-o`  | `/output` | Directory containing SnapRAID JSON files |
| `--log-format`     | `-l`  | `json`    | Log format (`json` or `text`)            |
| `--help`           | `-h`  |           | Show help and exit                       |
| `--version`        |       |           | Show version and exit                    |

## 📁 Expected File Structure

`go-snapraid-webui` expects your `go-snapraid` output JSON files in the `--output-dir`. File names should follow the pattern:

Example:

```
/output/
├── 2024-06-01T03:00:00Z.json
├── 2024-06-02T03:00:00Z.json
└── ...
```

These files are used to render the overview and details pages.

## 📄 License

MIT License. See [LICENSE](./LICENSE) for details.
