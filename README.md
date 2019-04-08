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
$ docker run -p 8080:8080 csmith/github-release-redirector -repo user/repo
```

## Options

```
Usage of github-release-redirector:
  -poll int
    	the amount of time to wait between polling for releases; 0 to disable polling (default 3600)
  -port int
    	the port to listen on for HTTP requests (default 8080)
  -redirect string
    	if specified, requests for / will be redirected to this url
  -repo string
    	the repository to redirect releases for, in user/repo format [required]
```

## Notes

If the `-redirect` option is specified, then requests to the root (i.e. `/`)
will be redirected to that URL, and no further processing will be done for
that request.

Releases are refreshed at startup and then, by default, once an hour. You can
customise this duration with the `-poll` option; setting the poll time to `0`
will disable polling entirely other than at startup.

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