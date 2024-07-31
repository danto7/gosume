function gosume -a profile -w gosume -d "wraps the gosume command integrate with the shell"
  set -l output ({{ .BinaryPath }} assume --print-secrets "$profile")

  
  if test "$status" -ne 0
    echo "$output"
    return $status
  end

  set -l env_whitelist AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_REGION AWS_SESSION_TOKEN
  for line in $output
    set -l pair (string split '=' $line)
    set -l key $pair[1]
    set -l value $pair[2]

    if contains $key $env_whitelist
      set -gx $key $value
    end
  end
end
