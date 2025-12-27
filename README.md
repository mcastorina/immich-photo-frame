# Digital Photo Frame

A desktop application to display photos pulled from an
[immich](https://immich.app/) server.


## Build

```bash
go build -o dpf .
```

## Configure

```bash
export IMMICH_API_ENDPOINT="http://immich:2283"
read -s IMMICH_API_KEY
export IMMICH_API_KEY
```

## Run

```bash
./dpf
```
