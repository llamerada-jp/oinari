# Oinari

Oinari(オイナリ) is the PoC project using distributed algorithms. In this program, by connecting users' browsers and collaborating with each other, it is possible to diffuse the data with little use of the server.

## Commands

Prepare keys for embedding and build this project.

```sh
$ cat secrets.json
{
    "cookie_key_pair": "<random base64 encoded key for cookie>",
    "github_client_id": "<client id of github oauth>",
    "github_client_secret": "<client secret of GitHub application>",
    "google_api_key": "<google api key of GitHub application>",
    "google_map_id": "<ID of google map>"
}

$ make setup
$ make build
```

## License

Apache License 2.0