# How to execute the examples

## Desktops

```sh
go run -tags=example $GOPATH/src/github.com/gabstv/ebiten/examples/rotate/main.go
```

## Web Browsers

```sh
gopherjs serve --tags=example
```

and access `http://127.0.0.1:8080/github.com/gabstv/ebiten/examples`.

## Android

Install [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile) first.

```sh
gomobile install -tags=example github.com/gabstv/ebiten/examples/rotate
```

## iOS

```sh
gomobile build -target=ios -tags=example -work github.com/gabstv/ebiten/examples/rotate
```

Then, open the `WORK` directory, open `main.xcodeproj`, add signing, and run the project.
