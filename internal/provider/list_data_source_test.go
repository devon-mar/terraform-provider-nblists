package provider

import (
	"fmt"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestListDataSource(t *testing.T) {
	var token string

	url := os.Getenv("TEST_NBLISTS_URL")
	t.Logf("Test url: %q", url)

	if url == "" {
		token = "abcdefghijklmnop"
		h := newTestListsHandler(t, token)
		h.addList("ip-addresses", map[string][]string{"tag": {"1"}}, []string{"192.0.2.1/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"2"}}, []string{})
		h.addList("ip-addresses", map[string][]string{"tag": {"3"}, "as_cidr": {"true"}}, []string{"192.0.2.3/32", "2001:db8::3/128"})
		h.addList("ip-addresses", map[string][]string{"tag": {"4"}, "as_cidr": {"false"}, "summarize": {"false"}}, []string{"192.0.2.4"})
		h.addList("prefixes", nil, []string{"192.0.2.0/26"})
		h.addList("ip-addresses", map[string][]string{"tag": {"5"}}, []string{"192.0.2.5/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"6"}}, []string{"192.0.2.6/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"7"}}, []string{"192.0.2.7/32"})
		h.addList("ip-addresses", map[string][]string{"parent": {"192.0.2.100/31"}, "summarize": {"true"}}, []string{"192.0.2.100/31"})
		h.addList("ip-addresses", map[string][]string{"parent": {"192.0.2.100/31"}, "summarize": {"false"}}, []string{"192.0.2.100/32", "192.0.2.101/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"10"}, "summarize": {"false"}}, []string{"192.0.2.10/32", "192.0.2.11/32"})
		s := httptest.NewServer(h)
		defer s.Close()
		url = s.URL
	} else {
		token = os.Getenv("TEST_NBLISTS_TOKEN")
	}

	providerConf := fmt.Sprintf(`
provider "nblists" {
    url = "%s"
	token = "%s"
}
`, url, token)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
// One IP
data "nblists_list" "one" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["1"] }
}

// Empty
data "nblists_list" "two" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["2"] }
}

// as_cidr=True
data "nblists_list" "three" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["3"] }
	as_cidr = true
	split_af = true
	no_cidr_single_ip = true
}

// as_cidr=False
data "nblists_list" "four" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["4"] }
	as_cidr = false
	summarize = false
}

// min
data "nblists_list" "five" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["5"] }
	min = 1
}

// max
data "nblists_list" "six" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["6"] }
	max = 1
}

// min and max
data "nblists_list" "seven" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["7"] }
	min = 1
	max = 1
}

// summarize=true
data "nblists_list" "eight" {
	endpoint = "ip-addresses"
	filter = { "parent" = ["192.0.2.100/31"] }
	summarize = true
}
// summarize=false
data "nblists_list" "nine" {
	endpoint = "ip-addresses"
	filter = { "parent" = ["192.0.2.100/31"] }
	summarize = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.nblists_list.one",
						"list.0",
						"192.0.2.1/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.one",
						"list.#",
						"1",
					),
					resource.TestCheckNoResourceAttr(
						"data.nblists_list.one",
						"list_no_cidr",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.two",
						"list.#",
						"0",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list.0",
						"192.0.2.3/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list.1",
						"2001:db8::3/128",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list_no_cidr.0",
						"192.0.2.3",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list_no_cidr.1",
						"2001:db8::3",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list_no_cidr.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list4.0",
						"192.0.2.3/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list4.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list6.0",
						"2001:db8::3/128",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three",
						"list6.#",
						"1",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.four",
						"list.0",
						"192.0.2.4",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.five",
						"list.0",
						"192.0.2.5/32",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.six",
						"list.0",
						"192.0.2.6/32",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.seven",
						"list.0",
						"192.0.2.7/32",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.eight",
						"list.0",
						"192.0.2.100/31",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.nine",
						"list.0",
						"192.0.2.100/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.nine",
						"list.1",
						"192.0.2.101/32",
					),
				),
			},
		},
	})

	// with allow_empty_filter=true
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "nblists" {
    url = "%s"
	token = "%s"
	allow_empty_filter = true
	request_timeout = 20
}

data "nblists_list" "test" {
	endpoint = "prefixes"
}
`, url, token),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.nblists_list.test",
						"list.0",
						"192.0.2.0/26",
					),
					resource.TestCheckNoResourceAttr(
						"data.nblists_list.test",
						"list.1",
					),
				),
			},
		},
	})

	// empty filter with allow_empty_filter=false
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
data "nblists_list" "test" {
	endpoint = "prefixes"
}
`,
				ExpectError: regexp.MustCompile(`Error getting list: filter is nil or empty`),
			},
		},
	})

	// min violated
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
data "nblists_list" "test" {
	endpoint = "ip-addresses"
	min = 2
	filter = {
		tag = ["6"]
	}
}
`,
				ExpectError: regexp.MustCompile(`The list has length \(1\) less than the min \(2\)`),
			},
		},
	})

	// max violated
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConf + `
data "nblists_list" "test" {
	endpoint = "ip-addresses"
	max = 1
	filter = {
		tag = ["10"]
	}
	summarize = false
}
`,
				ExpectError: regexp.MustCompile(`The list has length \(2\) greater than the max \(1\)`),
			},
		},
	})
}
