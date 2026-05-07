# gmcore-asset-mapper

GMCore SDK for asset mapping — bundles and SDKs register their public asset paths, and the application serves them under `/assets/`.

## Usage

```go
import (
    assetmapper "github.com/gmcorenet/sdk-gmcore-asset-mapper"
)

func init() {
    assetmapper.RegisterPath("crud-bundle", "./bundles/crud-bundle/public")
}

// In your main HTTP server:
http.Handle("/assets/", assetmapper.AssetHandler())

// In templates:
// {{ asset('bundles/crud-bundle/styles/crud.css') }}
// => /assets/bundles/crud-bundle/styles/crud.css
```

## Installing assets at startup

```go
if err := assetmapper.InstallAssets("./public"); err != nil {
    log.Fatal(err)
}
```

This copies all registered bundle asset directories into `public/assets/bundles/<name>/`.

## API

| Function | Description |
|----------|-------------|
| `RegisterPath(name, dir)` | Register a bundle/SDK public asset directory |
| `SetPublicDir(dir)` | Set application's public directory |
| `ResolveAsset(path)` | Resolve a template asset path to its URL |
| `AssetHandler()` | HTTP handler for serving `/assets/` paths |
| `InstallAssets(destDir)` | Copy all registered assets to destination |
| `AssetFunc()` | Returns `func(string)string` for template use |
| `AllPaths()` | Returns map of all registered paths |
