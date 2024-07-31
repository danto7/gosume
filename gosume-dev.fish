function gosume-dev
  go build -o bin/gosume cmd/gosume/main.go \
  && bin/gosume completion fish | source \
  && gosume $argv
end
