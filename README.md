# Usage:

```bash
speak C:\Temp\filename.txt
```

Audio output is saved to a temp file and played.

- On Linux, uses [play](https://linux.die.net/man/1/play)
- On Windows, uses [cmdmp3](https://github.com/jimlawless/cmdmp3)

Requires a `.env` file with a Speechify `API_KEY`.

