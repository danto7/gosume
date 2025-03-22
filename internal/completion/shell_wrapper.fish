function gosume -w gosume -d "wraps the gosume command integrate with the shell"
  set -l gosume (command -v gosume)
  $gosume --shell-wrapper $argv
  if test $status -eq "{{ .UpdateEnvironmentExitCode }}"
    set -l output ($gosume --shell-wrapper --print-secrets $argv)

    set -l env_whitelist AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_REGION AWS_SESSION_TOKEN
    for line in $output
      set -l pair (string split '=' $line)
      set -l key $pair[1]
      set -l value $pair[2]

      if contains $key $env_whitelist
        set -gx $key $value
      end
    end
    return 0
  else
    return $status
  end
end
