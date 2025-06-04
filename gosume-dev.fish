function gosume-dev-fish
  go build -o bin/gosume cmd/gosume/main.go \
  && bin/gosume completion fish | source \
  && PATH="./bin:$PATH" GOSUME_LOG_LEVEL=DEBUG gosume $argv
end
