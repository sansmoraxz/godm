# Godm

This is a file downloader in Go. Like age old [Internet Download Manager](https://www.internetdownloadmanager.com/) it will use a connection pool, split the file and download each chunk seperately.

Slight implementation differences:-

- Ranged fetch. It will download the file in multiple parts and then merge them together into a single file.
- Non resumable. It's still a work in progress. Resumable downloads will be added in the future.
- CLI. This is purely CLI app. No GUI as of yet.
- Using golang context is still a work in progress.
- If a chunk fails to download, it will retry downloading that chunk again.
