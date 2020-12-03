package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jncmaguire/release-notifier/internal/util"
)

type Ingester struct {
	TagName string `json:"tag_name"`
}

func (c *Client) request(method string, path string, pathArgs map[string]interface{}, body interface{}) ([]byte, error) {
	request, err := util.BuildRequest(method, c.APIURL, path, pathArgs, body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer: %s", c.APIToken))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(`Accept`, `application/vnd.github.v3+json`)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

func (c *Client) getReleases(owner string, repo string, perPage int, page int) ([]util.Release, error) {
	data, err := c.request(http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases", owner, repo), map[string]interface{}{
		`per_page`: perPage,
		`page`:     page,
	}, nil)

	if err != nil {
		return []util.Release{}, err
	}

	objects := make([]Ingester, perPage)

	if err = json.Unmarshal(data, &objects); err != nil {
		return []util.Release{}, fmt.Errorf("%w: %v", err, string(data))
	}

	releases := make([]util.Release, 0, perPage)

	i := 0
	for i < perPage {
		release, err := util.NewReleaseFromString(objects[i].TagName)

		if err == nil {
			releases = append(releases, release)
		}
	}

	return releases, err
}
