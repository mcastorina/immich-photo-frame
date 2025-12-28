# Immich Photo Frame

A desktop application to display photos pulled from an
[immich](https://immich.app/) server.


## Build

```bash
go build -o ipf .
```

## Configure

The application is configured via TOML. By default it looks for a `config.toml`
file in current working directory, however if the `IMMICH_PHOTO_FRAME_CONFIG`
environment variable is set, it uses that.

```toml
[immich]
immichAPIEndpoint = "http://immich:2283"
immichAPIKey = "consider using IMMICH_API_KEY env var"

[localStorage]
useLocalStorage  = false
localStorageSize = "1337"
localStoragePath = ""

[inMemoryCache]
useInMemoryCache  = true
inMemoryCacheSize = "512 MB"
```

### Immich

The `immich` section configures connecting to the immich server. These values
can also be configured via environment variables, which take precedence and are
generally considered more secure.

| key | env | type | description |
| --- | --- | --- | --- |
| `immichAPIEndpoint` | `IMMICH_API_ENDPOINT` | string | URL of the immich API endpoint to use |
| `immichAPIKey` | `IMMICH_API_KEY` | string | API key for [authenticating to the immich server](https://api.immich.app/authentication) |

### Local Storage

The `localStorage` section configures where and how many assets to save to
local disk. This feature allows storing and loading assets locally if your
immich server is not always available.

| key | type | description |
| --- | --- | --- |
| `useLocalStorage` | bool | Enable storing assets locally |
| `localStorageSize` | string | Amount of bytes to use for storing assets (in human-readable text) |
| `localStoragePath` | string | Path to store assets |

### In-Memory Cache

The `inMemoryCache` section configures how much memory to allocate before going
to disk or the immich server when loading assets.

| key | type | description |
| --- | --- | --- |
| `useInMemoryCache` | bool | Enable storing assets in-memory |
| `inMemoryCacheSize` | string | Amount of bytes to use for storing assets (in human-readable text) |

## Run

```bash
./ipf
```
