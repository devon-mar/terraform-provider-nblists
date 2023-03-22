package provider

import (
	"fmt"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestListDataSource(t *testing.T) {
	var token string

	url := os.Getenv("TEST_NBLISTS_URL")

	if url == "" {
		token = "abcdefghijklmnop"
		h := newTestListsHandler(t, token)
		h.addList("ip-addresses", map[string][]string{"tag": {"1"}}, []string{"192.0.2.1/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"2"}}, []string{})
		h.addList("ip-addresses", map[string][]string{"tag": {"3"}, "as_cidr": {"true"}}, []string{"192.0.2.3/32", "2001:db8::3/128"})
		h.addList("ip-addresses", map[string][]string{"tag": {"3"}, "family": {"4"}}, []string{"192.0.2.3/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"3"}, "family": {"6"}}, []string{"2001:db8::3/128"})
		h.addList("ip-addresses", map[string][]string{"tag": {"4"}, "as_cidr": {"false"}, "summarize": {"false"}}, []string{"192.0.2.4"})
		h.addList("aggregates", nil, []string{"192.0.2.0/24"})
		h.addList("ip-addresses", map[string][]string{"tag": {"5"}}, []string{"192.0.2.5/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"6"}}, []string{"192.0.2.6/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"7"}}, []string{"192.0.2.7/32"})
		h.addList("ip-addresses", map[string][]string{"parent": {"192.0.2.100/31"}, "summarize": {"true"}}, []string{"192.0.2.100/31"})
		h.addList("ip-addresses", map[string][]string{"parent": {"192.0.2.100/31"}, "summarize": {"false"}}, []string{"192.0.2.100/32", "192.0.2.101/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"10"}, "summarize": {"false"}}, []string{"192.0.2.10/32", "192.0.2.11/32"})
		h.addList("ip-addresses", map[string][]string{"tag": {"11"}, "summarize": {"false"}, "as_cidr": {"false"}}, []string{"192.0.2.12", "2001:db8::12"})
		// mixed single IPs with prefixes
		h.addList(
			"prefixes",
			map[string][]string{"tag": {"p1"}, "summarize": {"false"}},
			[]string{"192.0.2.0/27", "192.0.2.200/32", "2001:db8::/64", "2001:db8::200/128"},
		)
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

// family=4
data "nblists_list" "three4" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["3"] }
	family = 4
}

// family=6
data "nblists_list" "three6" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["3"] }
	family = 6
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

// mixed single IP prefixes and regular ones
data "nblists_list" "ten" {
	endpoint = "prefixes"
	filter = { "tag" = ["p1"] }
	summarize = false
	split_af = true
	no_cidr_single_ip = true
}

// split af on single IPs (no CIDR)
data "nblists_list" "eleven" {
	endpoint = "ip-addresses"
	filter = { "tag" = ["11"] }
	summarize = false
	as_cidr = false
	split_af = true
	no_cidr_single_ip = true
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
						"data.nblists_list.three4",
						"list.0",
						"192.0.2.3/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three4",
						"list.#",
						"1",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.three6",
						"list.0",
						"2001:db8::3/128",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.three6",
						"list.#",
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

					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list.0",
						"192.0.2.0/27",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list.1",
						"192.0.2.200/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list.2",
						"2001:db8::/64",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list.3",
						"2001:db8::200/128",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list_no_cidr.0",
						"192.0.2.0/27",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list_no_cidr.1",
						"192.0.2.200",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list_no_cidr.2",
						"2001:db8::/64",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list_no_cidr.3",
						"2001:db8::200",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list_no_cidr.#",
						"4",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list4.0",
						"192.0.2.0/27",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list4.1",
						"192.0.2.200/32",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list4.#",
						"2",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list6.0",
						"2001:db8::/64",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list6.1",
						"2001:db8::200/128",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.ten",
						"list6.#",
						"2",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list.0",
						"192.0.2.12",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list.1",
						"2001:db8::12",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list.#",
						"2",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list_no_cidr.0",
						"192.0.2.12",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list_no_cidr.1",
						"2001:db8::12",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list_no_cidr.#",
						"2",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list4.0",
						"192.0.2.12",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list4.#",
						"1",
					),

					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list6.0",
						"2001:db8::12",
					),
					resource.TestCheckResourceAttr(
						"data.nblists_list.eleven",
						"list6.#",
						"1",
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
	endpoint = "aggregates"
}
`, url, token),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.nblists_list.test",
						"list.0",
						"192.0.2.0/24",
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
	endpoint = "aggregates"
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
