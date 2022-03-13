# Oinari

Oinari(オイナリ) is the PoC project using distributed algorithms. In this program, by connecting users' browsers and collaborating with each other, it is possible to diffuse the data with little use of the server.

## Commands

Prepare keys for embedding and build this project.

```sh
$ cat keys.json
{
    "google_analytics_tracking_id": "<key of google analytics>",
    "google_api_key": "<google api key>",
    "google_map_id": "<ID of google map>"
}

$ make setup
$ make build
```

## License

Apache License 2.0