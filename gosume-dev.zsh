# TODO: For PROD we need an install script and different names for the functions (too often gosume!)
function gosume-dev-zsh() {
    go build -o bin/gosume cmd/gosume/main.go \
    && source <(bin/gosume completion zsh) \
    && gosume "$@"
}
