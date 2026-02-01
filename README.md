# Immich Photo Frame

A desktop application to display photos pulled from an
[immich](https://immich.app/) server.



## Configure

The application is configured via TOML. By default it looks for a `config.toml`
file in the current working directory, however if the
`IMMICH_PHOTO_FRAME_CONFIG` environment variable is set, it uses that.

```toml
[app]
immichAlbums = ["Photo Frame"]
imageDelay = "10s"

[remote]
immichAPIEndpoint = "http://immich:2283"
immichAPIKey = "consider using IMMICH_API_KEY env var"

[localStorage]
useLocalStorage  = true
localStorageSize = "512 MB"
localStoragePath = "$HOME/.ipf"
```

### App

The `app` section configures how the application will run.

| key | type | default | description |
| --- | --- | --- | --- |
| `immichAlbums` | []string | All albums found | List of the immich albums to use |
| `imageDelay` | string | `5s` | Amount of time between displaying images (in human-readable text) |
| `imageScale` | float | `1` | Value between 0 and 1 for scaling the image (higher values for better resolution) |
| `historySize` | int | `10` | How many images to keep for going backwards |
| `planAlgorithm` | string | `sequential` | Algorithm for advancing through configured albums and assets |
| `immichAlbumRefreshInterval` | string | `24h` | Amount of time before checking the immich server for new albums and assets (in human-readable text) |
| `imageText` | []string | `["image-date-time:20", "image-location:16"]` | Text configuration to display on-screen |

#### Plan Algorithms

Plan algorithms define how to advance through configured albums and assets.
Below is a list of the existing algorithms followed by a short description.

* **sequential:** Albums are iterated in the order they are configured and
  assets are iterated in the order the immich server specifies.
* **shuffle:** All assets from all albums are shuffled and shown once before
  shuffling again.

#### Text Configuration

Text is configured via the `imageText` attribute as an array of strings. The
order of the strings represents the line from top to bottom. The string expects
a format of `"name:size"`, where `name` is the name of a text formatter and
`size` is the `float32` text font size, which defaults to `16` if not
specified. Below is a list of existing text formatters followed by a short
description.

* **image-date-time:** Date and time when the image was taken, according to its
  EXIF data, with some custom formatting:
  * If the date is within 1 week of the current time, display the weekday and
    time.
  * If the date is within 3 months of the current time, display a relative
    offset.
  * Otherwise, display just the date.
* **image-location:** Location where the image was taken, according to its EXIF
  data.

### Remote

The `remote` section configures connecting to the immich server. These values
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
| `localStoragePath` | string | Absolute path to store assets (supports environment variables) |

### In-Memory Cache

The `inMemoryCache` section configures how much memory to allocate before going
to disk or the immich server when loading assets. This typically is only useful
if you have a lot of RAM and not a lot of storage.

| key | type | description |
| --- | --- | --- |
| `useInMemoryCache` | bool | Enable storing assets in-memory |
| `inMemoryCacheSize` | string | Amount of bytes to use for storing assets (in human-readable text) |


## Development

This project uses [fyne](https://fyne.io), a cross-platform GUI framework.
Theoretically you can build and run this "anywhere," but it has only been
designed and tested on MacOS and a Raspberry Pi.

To build and run for local development:

```bash
go build -o ipf .
./ipf --debug
```

## Raspberry Pi

The intended final destination for this project is to run on a Raspberry Pi,
which has limited RAM and storage.

The application running on a RPi takes roughly 350M of RAM on average, with
spikes as high as 600M. My preliminary attempts to get it work on a RPi Zero 2W
were met with OOM kills. Newer models with at least 1GB of RAM are recommended.

| Hardware | Minimum required |
| --- | --- |
| RAM | 1 GB |
| SD Card | 8 GB |


### Build and run (remotely)

```bash
sudo apt install    \
    golang          \
    libxrandr-dev   \
    libxinerama-dev \
    libxi-dev       \
    libxxf86vm-dev  \
    libxcursor-dev

go build -o ipf .

DISPLAY=:0 ./ipf
```
