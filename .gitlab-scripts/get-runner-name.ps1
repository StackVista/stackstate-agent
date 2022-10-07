echo $($(cat C:\gitlab-runner\config.toml | grep token | awk '{print $3}') -replace '"',"").Substring(0,8)
