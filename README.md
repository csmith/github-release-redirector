# github-release-redirector

Small go application that listens for incoming HTTP requests and redirects
to assets from the latest release published to a GitHub repository.

## Usage

```console
$ go install .
$ github-release-redirector -port 1234 -repo user/repo
```

or using docker:

```console
$ docker run -p 8080 csmith/github-release-redirector -repo user/repo
```

## Notes

Releases are refreshed at startup and then once an hour.

Assets are matched on their name, so an asset attached to the latest release
named "myproject-installer-1.2.3.exe" will be available at the URL
`/myproject-installer-1.2.3.exe`.

Any url not mapped to an asset is responded to with a 404 not found error;
if there is no latest release available all requests will respond with 500
internal server error.

The service listens only on HTTP, not HTTPS. For production use it should
be placed behind an SSL-terminating proxy such as Traefik, Nginx or HAProxy.

## Licence

This software is released under the MIT licence. See the LICENCE file for
full details.