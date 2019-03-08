package coding

import (
	"github.com/drone/go-scm/scm"
	"net/url"
	"strconv"
)

func encodeListOptions(opts scm.ListOptions) string {
	params := url.Values{}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("pageSize", strconv.Itoa(opts.Size))
	}
	return params.Encode()
}
