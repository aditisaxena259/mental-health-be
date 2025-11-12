# Mental Health BE — Cloudinary Uploads

This backend stores JPEG complaint/apology attachments in Cloudinary when configured, with a safe local fallback under `./uploads` when not.

## Environment

Create `.env` (already git-ignored) with either separate vars or a single URL:

```
CLOUDINARY_CLOUD_NAME=dioalthuc
CLOUDINARY_API_KEY=YOUR_KEY
CLOUDINARY_API_SECRET=YOUR_SECRET
# OR
CLOUDINARY_URL=cloudinary://<api_key>:<api_secret>@<cloud_name>
```

Other env like `DATABASE_URL` are used by the app; see `.env` in the repo for examples.

## Run

Start the server (defaults to http://localhost:8080):

```
go run main.go
```

Static fallback files are served from `/uploads` when Cloudinary is not available.

## Health check

Public endpoint to verify Cloudinary configuration:

```
GET /api/health/cloudinary -> { ok: true } if SDK initializes
```

## Test the flow

1. Login as a student to get a token.
2. Create a complaint with a JPEG attachment using multipart key `attachments`.
3. Fetch complaints:
   - Cloudinary configured: `attachments[].FileURL` is an HTTPS Cloudinary URL and `PublicID` non-empty.
   - Not configured: `attachments[].FileURL` is a local `/uploads/...` path.
4. Admin/chief admin fetches `/api/admin/complaints` (needs block assignment for plain admin) to view same attachment metadata.

## Delete flow

Deleting a complaint attempts Cloudinary destroy using `PublicID` and removes the local directory fallback if present.

## Apologies

Apology JPEG attachments behave identically via `ApologyAttachment` records.

## Frontend Usage

Display the image by using the `FileURL`. If `PublicID` exists, you can later build transformed URLs on the frontend (e.g., resize) using Cloudinary JS SDK. If `PublicID` empty, treat `FileURL` as static served asset.

## Security Notes

Never commit `.env`. API secret must remain server-side; only `FileURL` and `PublicID` are returned to clients.
