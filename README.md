# Video Processing

![Architecture Diagram](./vid-pro.webp)

Video transcoding pipeline. Upload a video → R2 storage → RabbitMQ job queue → worker transcodes to 360p/480p/720p/1080p via ffmpeg → uploads renditions back to R2.

## Dependencies

- [Go](https://go.dev) 1.25+
- [Docker](https://docker.com)
- [ffmpeg](https://ffmpeg.org) (for worker)

## Quick Start

### 1. PostgreSQL & RabbitMQ

```sh
docker run -d --name vid-pro-db -e POSTGRES_PASSWORD=admin -p 5432:5432 postgres
docker run -d --name vid-pro -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

### 2. Database

```sh
# Create the database inside the container
docker exec vid-pro-db psql -U postgres -c "CREATE DATABASE \"vid-pro-db\";"

# Run migrations
./Scripts/migrate.sh up
```

### 3. Env

Copy the following into `.env`:

```env
PORT=3001
DATABASE_URL=postgresql://postgres:admin@localhost:5432/vid-pro-db?sslmode=disable
RABBIT_MQ_URL=amqp://guest:guest@localhost:5672/
R2_ACCOUNT_ID=<your-r2-account-id>
R2_ACCESS_KEY_ID=<your-r2-key>
R2_ACCESS_KEY_SECRET=<your-r2-secret>
R2_PUBLIC_DOMAIN=https://pub-<hash>.r2.dev
BUCKET_NAME=vid-proccessing
```

### 4. Run

```sh
go run ./cmd
```

## API

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/videos` | Upload a video |
| `GET` | `/api/v1/videos` | List all videos |
| `GET` | `/api/v1/videos/:id` | Get video by ID |

### Create a video

```sh
curl -X POST http://localhost:3001/api/v1/videos \
  -F "title=My Video" \
  -F "file=@/path/to/video.mp4"
```

Optional `video_id` field — auto-generated as UUID if omitted.

## Architecture

1. **POST /videos** — saves uploaded file to temp, uploads to R2, inserts DB record with `uploaded` status, publishes job ID to RabbitMQ
2. **Worker** — consumes jobs, downloads from R2, transcodes with ffmpeg at 4 resolutions, uploads renditions to R2, updates DB status to `ready`
3. **GET /videos** — reads from PostgreSQL

### Queue retries

Failed jobs are retried up to 3 times (`x-delivery-count`), then discarded. Manual acks — messages requeue on transient failure and stay in-flight if the worker crashes.
