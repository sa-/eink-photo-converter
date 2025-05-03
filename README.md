# eink-photo-converter

A better converter for the Waveshare Photo Painter (B)

Install
```sh
go install github.com/sa-/eink-photo-converter@latest
```

This installs to `~/go/bin/`, make sure it's on your PATH.

Usage example

```sh
eink-photo-converter --in-dir ~/Downloads/some-photos --out-dir $(mktemp -d)
```

Help text
```sh
> eink-photo-converter --help 
Usage of eink-photo-converter:
  -in-dir string
        Input directory containing images
  -out-dir string
        Output directory for converted images
```
