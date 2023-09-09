# Documents of Oinari

Those files are configured for Hugo Extended.

## Read the document in your local machine

To install Hugo Extended.
See [the official document](https://gohugo.io/installation/linux/#build-from-source) to know details.
```console
CGO_ENABLED=1 go install -tags extended github.com/gohugoio/hugo@latest
```

Start Hugo server.
```console
# at this dir
hugo serve
```

And access described url like http://127.0.0.1:1313/ using your browser.

## Add new a new page

Add an entry linking to a new page at `docs/content/menu/index.md` referring to existing writing.

## Refract changes for the online document

TODO: put online document 