# Yankd IPC

Yankd provides an Inter-Process Communication (IPC) mechanism via an HTTP server
listening on a Unix Domain Socket. This allows external clients or GUI wrappers
to interact with the daemon to query, search, and manage clipboard history.

## Unix Socket

The socket file is named `yankd.sock` and is located in the directory specified
by the `$XDG_RUNTIME_DIR` environment variable. If `$XDG_RUNTIME_DIR` is not
set, it defaults to `/tmp`.

Clients must connect to this Unix socket and speak standard HTTP/1.1. In tools
like `curl`, this is done using the `--unix-socket` flag.

## Endpoints

All responses are generally returned with the `Content-Type: application/json`
header unless specified otherwise. In case of errors, the response will be plain
text with an appropriate HTTP status code.

### `GET /get/{id}`

Retrieves a specific clipboard event by its UUID.

- **Parameters:** `id` (UUID in the path)
- **Response:** JSON representation of the `ClipboardEvent`.

### `POST /get`

Retrieves multiple clipboard events by their UUIDs.

- **Body:** JSON array of UUID strings (e.g., `["uuid-1", "uuid-2"]`).
- **Response:** JSON array of `ClipboardEvent` objects.

### `POST /set/{id}`

Restores a clipboard event to the system clipboard using its UUID. This
retrieves the event from the database and sets it as the active clipboard
content.

- **Parameters:** `id` (UUID in the path)
- **Response:** JSON representation of the restored `ClipboardEvent`.

### `GET /search`

Searches for clipboard events in the database using advanced query strings.

- **Query Parameters:**
  - `query` (optional string): The search string. Follows the advanced query
    format (e.g., `type:image`, exact match, time filter). If empty, returns the
    most recent clipboard events.
  - `limit` (optional integer): The maximum number of results to return.
- **Response:** JSON array of search results.

### `POST /delete`

Deletes one or more clipboard events from the database.

- **Body:** JSON array of UUID strings to delete.
- **Response:** JSON integer representing the number of events deleted.

### `POST /wipe`

Wipes the entire clipboard history database.

- **Response:** JSON integer representing the number of events deleted.

### `POST /set-custom`

Sets arbitrary clipboard content with one or more MIME types. This allows clients
to inject custom clipboard entries into the history.

- **Body:** JSON object with the following structure:
  ```json
  {
    "mime_type": "text/plain",  // Optional: primary MIME type (auto-detected if omitted)
    "preview": "hello world",   // Optional: preview text (auto-generated if omitted)
    "entries": [
      {
        "mime_type": "text/plain",
        "blob": "aGVsbG8="  // Base64 encoded content
      }
    ]
  }
  ```
- **Response:** JSON object `{"status": "ok"}` on success.

**Example using curl:**

```bash
curl --unix-socket $XDG_RUNTIME_DIR/yankd.sock -X POST http://localhost/set-custom \
  -H "Content-Type: application/json" \
  -d '{"entries":[{"mime_type":"text/plain","blob":"aGVsbG8gd29ybGQ="}]}'
```

### `/pause/{state}`

- **Parameters:** `state` (integer in the path):
  - `-1`: Unpause (resume monitoring)
  - `0`: Toggle current state
  - `1`: Pause (stop monitoring)
- **Response:** JSON boolean indicating the new paused state (`true` if paused,
  `false` otherwise).

### `POST /echo`

Echoes the request body back to the client. Useful for testing the connection or
pinging the daemon.

- **Body:** Any plain text string.
- **Response:** Plain text matching the request body.

## Example

Using `curl` to list the latest 10 clipboard events (assuming `$XDG_RUNTIME_DIR`
is `/run/user/1000`):

```bash
curl --unix-socket /run/user/1000/yankd.sock "http://localhost/search?limit=10"
```
