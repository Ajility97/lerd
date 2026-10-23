# Importing from Laravel Sail

The **sail** plugin migrates a Laravel Sail project's data into lerd, database and S3/MinIO storage included. For the wider picture, why you would move and what changes day to day, see the [Laravel Sail alternative](../getting-started/sail.md) guide.

```bash
lerd plugin install sail   # once
cd ~/Projects/myapp        # the Sail project root
lerd link && lerd env      # the site and its lerd database
lerd sail:import
```

The import is also on the site's Commands menu in the dashboard as **Import from Sail**, and AI assistants get it as `sail_import`. The plugin's Documentation tab in the dashboard has its full instructions.

`lerd sail ...` itself still runs the project's own `vendor/bin/sail` on the host, so `lerd sail up` and `lerd sail artisan migrate` keep working.

---

## Prerequisites

- **Docker** (Docker Desktop, Docker Engine) **or Podman** with the `compose` plugin, whichever you used to run Sail
- The Sail project must have a `docker-compose.yml` in its root
- The project linked with `lerd link` and set up with `lerd env`, so the site has a lerd database to import into

---

## What it does

1. **Reads the compose file** with Sail's original `.env.before_lerd` and moves any port lerd already uses 20000 up (3306 becomes 23306), so Sail can start alongside lerd without clashes
2. **Leaves out the ports** of app and proxy services, which the import never starts
3. **Starts only the data services** (the database, and MinIO when the project stores files on S3) with `compose up -d --no-deps`, so Sail's app container is never built and a broken app image cannot block the import
4. **Waits for the database**, then finds the database to import: the only one Sail holds, or the one the project's env names
5. **Dumps it** inside the Sail container and imports it with `lerd db:import --fresh`, after a `before-sail-import` snapshot that `lerd db:restore` brings back
6. **Copies MinIO's files** into the site's lerd bucket with `lerd s3:import`
7. **Stops Sail** again, unless `--no-stop` is passed

---

## Options

| Option | Description |
|---|---|
| `--db-name <name>` | The Sail database, when Sail holds several |
| `--db-user <user>`, `--db-password <pass>` | Sail's database login, when it is not the one in the compose file |
| `--bucket <name>` | The MinIO bucket to copy, when it is not `AWS_BUCKET` |
| `--skip-s3` | Leave MinIO's files alone |
| `--no-stop` | Leave Sail running after the import |
| `--yes` | Don't ask before replacing the lerd database |

---

## Credentials

`lerd env` keeps the project's original `.env` as `.env.before_lerd` before it writes lerd's values. The import reads Sail's settings from that copy, so the compose file resolves to the database login, bucket and S3 settings the Sail containers were created with, even after `.env` points at lerd. MinIO's login comes from `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` in the compose file.

---

## Copying S3 files from anywhere

The file copy is lerd's own `lerd s3:import`, which works with any S3 server, not only Sail's MinIO:

```bash
lerd s3:import --endpoint 127.0.0.1:9000 --access-key key --secret-key secret --bucket uploads --to myapp
```

It copies every object of the bucket into the lerd bucket named by `--to`, creating it when missing. Objects with the same key are overwritten and nothing is deleted. Give the endpoint as `https://host:port` for a server that speaks HTTPS; a bare `host:port` or `http://` is plain HTTP. With `--bucket` it reads that bucket directly, so a key that may read it but not list the server's buckets still works; leave it out when the server holds a single bucket.

---

## Troubleshooting

**`docker compose` / `podman compose` not found**

Install Docker Desktop, Podman Desktop, or the relevant compose plugin and ensure it is on your `PATH`.

**The database never answers**

The Sail volume was created with a login the compose file no longer says. Pass it:

```bash
lerd sail:import --db-user root --db-password secret
```

**Sail holds several databases**

```bash
lerd sail:import --db-name myapp
```

**The S3 copy fails**

Import the database alone with `--skip-s3`, then copy the files with `lerd s3:import`.
