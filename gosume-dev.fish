function gosume-dev
  go build -o bin/gosume cmd/gosume/main.go \
  && bin/gosume completion fish | source \
  && GOSUME_LOG_LEVEL=DEBUG bin/gosume $argv
end
