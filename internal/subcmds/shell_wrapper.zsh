function gosume() {
    if ! {{ .BinaryPath }} --shell-wrapper $@; then
        if [[ $? -eq "{{ .UpdateEnvironmentExitCode }}" ]] ; then
            output=$({{ .BinaryPath }} --shell-wrapper --print-secrets $@)

            IFS=$'='
            echo "$output" | while read -r key value; do
                if [[ $key == "AWS_ACCESS_KEY_ID" || $key == "AWS_SECRET_ACCESS_KEY" || $key == "AWS_REGION" || $key == "AWS_SESSION_TOKEN" ]] ; then
                    export $key=$value
                fi
            done
            return 0
        else
            echo "ERROR: Failed to update environment variables"
            return $?
        fi
    fi
}
