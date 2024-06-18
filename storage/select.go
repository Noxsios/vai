package storage

import (
	"fmt"
	"net/url"

	"github.com/package-url/packageurl-go"
	"github.com/spf13/afero"
)

func SelectFetcher(uri, previous *url.URL) (Fetcher, error) {
	switch uri.Scheme {
	case "http", "https":
		return NewHTTPFetcher(), nil
	case "pkg":
		pURL, err := packageurl.FromString(uri.String())
		if err != nil {
			return nil, err
		}

		switch pURL.Type {
		case "github":
			return NewGitHubClient(), nil
		case "gitlab":
			client, err := NewGitLabClient(pURL.Qualifiers.Map()["base"])
			if err != nil {
				return nil, err
			}
			return client, nil
		default:
			return nil, fmt.Errorf("unsupported type: %q", pURL.Type)
		}
	case "file":
		switch previous.Scheme {
		case "file":
			return NewLocalFetcher(afero.NewOsFs()), nil
		case "http", "https":
			return NewHTTPFetcher(), nil
		case "pkg":
			pURL, err := packageurl.FromString(previous.String())
			if err != nil {
				return nil, err
			}
			switch pURL.Type {
			case "github":
				return NewGitHubClient(), nil
			case "gitlab":
				client, err := NewGitLabClient(pURL.Qualifiers.Map()["base"])
				if err != nil {
					return nil, err
				}
				return client, nil
			default:
				return nil, fmt.Errorf("unsupported type: %q", pURL.Type)
			}
		default:
			return nil, fmt.Errorf("unsupported scheme: %q", previous.Scheme)
		}
	default:
		return nil, fmt.Errorf("unsupported scheme: %q", uri.Scheme)
	}
}
