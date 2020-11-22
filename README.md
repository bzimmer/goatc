### Overview

The goal of this tool is to provide simple access to the [goatcounter](https://www.goatcounter.com) API from the command line. Currently only the exporting of visit data is available.

### Examples

- To iterate through all sites and summarize the number of visits from non-bots:
```sh
for site in $(goatc sites | jq -r '. | join(" ")'); do
  goatc visits $site | jq --arg site $site '.stats | map(select(.bot == 0)) | group_by(.path) | map ({path:.[0].path, count:length}) | sort_by(-.count) | {"site":$site, "visits":.}'
done
```

### Goatcounter API documentation for reference
- [documentation](https://www.goatcounter.com/api)
- [model](https://www.goatcounter.com/api.html)
