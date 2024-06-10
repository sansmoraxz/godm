# Godm

This is a file downloader in Go. Like age old [Internet Download Manager](https://www.internetdownloadmanager.com/) it will use a connection pool, split the file and download each chunk seperately.

Slight implementation differences:-

- Ranged fetch. It will download the file in multiple parts and then merge them together into a single file.
- No network compression. Since we are downloading to the same file it's near to pull it off with compression. Compression will always need another stage to decompress the downloaded payload, which is only possible after the full file has been downloaded.
- CLI. This is purely CLI app. No GUI as of yet.
