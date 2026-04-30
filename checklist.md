lightcloud - Mid-Presentation Code Review Checklist
=====================================================

## Security
- SQL injection: use prepared statements, never concatenate raw strings into queries
- Path traversal: apply filepath.Clean() and validate upload paths don't escape the uploads/ directory
- Session ID generation: use crypto/rand, not math/rand
- File upload: enforce file size limits on r.Body; validate MIME type by content, not just extension

## Authentication & Session
- Every handler that requires login must validate the session before doing anything
- On logout, delete the session from the DB (not just the cookie)

## Error Handling
- Unify the if err != nil pattern and error message format across all handlers
- Return appropriate HTTP status codes on errors — avoid sending 200 for everything
- Distinguish log.Fatal (unrecoverable) vs log.Println (recoverable) correctly

## Code Structure
- Handlers should only deal with request/response — move DB logic into the db/ package
- Verify model/ structs match the actual DB schema
- Extract duplicated code into common.go

## HTTP & Routing
- Check r.Method branching on all endpoints (GET vs POST)
- No hardcoded values (port, paths, secrets) — move to constants or a config

## Database
- Confirm defer db.Close() is placed correctly
- Add transactions where multiple DB writes need to be atomic

## Stability
- Hunt for potential nil dereferences that could cause a panic during the demo